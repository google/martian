// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cache enables caching and replaying HTTP responses.
package cache

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/martian"
	"github.com/google/martian/log"
	"github.com/google/martian/parse"
)

const (
	// CustomKey is the context key for setting a custom cache key for a request.
	// This can be used by any upstream modifiers to set a custom cache key via the context.
	CustomKey = "cache.CustomKey"

	// cachedResponseCtxKey is the context key for storing the cached response for a request.
	cachedResponseCtxKey = "cache.Response"

	// defaultBucket is the default bucket name for boltdb.
	defaultBucket = "martian.cache"

	// defaultFileOpTimeout is the default timeout when doing file operations, e.g. open.
	defaultFileOpTimeout = 10 * time.Second
)

func init() {
	parse.Register("cache.Modifier", modifierFromJSON)
}

type Modifier struct {
	db *bolt.DB
	// Bucket is the name of the database bucket.
	Bucket string
	// Update determines whether the cache will be updated with live responses.
	Update bool
	// Replay determines whether the modifier will replay cached responses.
	Replay bool
	// Hermetic determines whether to prevent request forwarding on cache miss.
	Hermetic bool
}

type modifierJSON struct {
	File     string               `json:"file"`
	Bucket   string               `json:"bucket"`
	Update   bool                 `json:"update"`
	Replay   bool                 `json:"replay"`
	Hermetic bool                 `json:"hermetic"`
	Scope    []parse.ModifierType `json:"scope"`
}

// NewModifier returns a cache and replay modifier.
// The returned modifier will be in non-hermetic passthrough mode using a default bucket name.
// `filepath` is the filepath to the boltdb file containing cached responses.
func NewModifier(filepath string) (*Modifier, error) {
	log.Infof("cache.Modifier: opening boltdb file %q", filepath)
	db, err := bolt.Open(filepath, 0600, &bolt.Options{
		Timeout: defaultFileOpTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("cache.Modifier: bolt.Open(%q): %v", filepath, err)
	}
	runtime.SetFinalizer(db, func(db *bolt.DB) {
		log.Infof("cache.Modifier: closing boltdb file %q", db.Path())
		if err := db.Close(); err != nil {
			log.Errorf("cache.Modifier: db.Close(): %v", err)
		}
	})

	return &Modifier{
		db:       db,
		Bucket:   defaultBucket,
		Update:   false,
		Replay:   false,
		Hermetic: false,
	}, nil
}

// Close closes the underlying database file.
func (m *Modifier) Close() error {
	if m.db != nil {
		runtime.SetFinalizer(m.db, nil)
		return m.db.Close()
	}
	return nil
}

// modifierFromJSON parses JSON into cache.Modifier.
//
// Example JSON Configuration message:
// {
//   "file":     "/some/cache.db",
//   "bucket":   "some_project",
//   "update":   true,
//   "replay":   true,
//   "hermetic": false
// }
//
// `file` is the filepath to the cache database file.
// `bucket` is the name of the boltdb bucket to use. It will be created in the db on cache update if it doesn't already exist.
// If `update` is true, the database will be updated with live responses, e.g. on cache miss or when not replaying.
// If `replay` is true, the modifier will attempt to replay responses from its cache.
// If `hermetic` is true, the modifier will return error if it cannot replay a cached response, e.g. on cache miss or not replaying.
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, fmt.Errorf("cache.Modifier: json.Unmarshal(): %v", err)
	}

	mod, err := NewModifier(msg.File)
	if err != nil {
		return nil, err
	}
	if msg.Bucket != "" {
		mod.Bucket = msg.Bucket
	}
	mod.Update = msg.Update
	mod.Replay = msg.Replay
	mod.Hermetic = msg.Hermetic

	log.Infof("cache.Modifier: created modifier from JSON: %v", mod)
	return parse.NewResult(mod, msg.Scope)
}

// String returns a string representation of the modifier.
func (m *Modifier) String() string {
	return fmt.Sprintf("cache.Modifier{File: %q, Bucket: %q, Update: %t, Replay: %t, Hermetic: %t}",
		m.db.Path(), m.Bucket, m.Update, m.Replay, m.Hermetic)
}

// ModifyRequest performs a cache lookup or passthrough for the request.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	if !m.Replay {
		if m.Hermetic {
			return errors.New("cache.Modifier.ModifyRequest(): in hermetic mode and not replaying from cache")
		}
		return nil
	}

	key, err := getCacheKey(req)
	if err != nil {
		return fmt.Errorf("cache.Modifier.ModifyRequest(): getCacheKey(): %v", err)
	}

	return m.db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(m.Bucket)); b != nil {
			if cached := b.Get(key); cached != nil {
				res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(cached)), req)
				if err != nil {
					return fmt.Errorf("cache.Modifier.ModifyRequest(): http.ReadResponse(): %v", err)
				}
				ctx := martian.NewContext(req)
				ctx.SkipRoundTrip()
				ctx.Set(cachedResponseCtxKey, res)
				return nil
			}
		}
		if m.Hermetic {
			return errors.New("cache.Modifier.ModifyRequest(): in hermetic mode and no cached response found")
		}
		return nil
	})
}

// ModifyResponse applies the cached response, updates the cache with real response, or just passes through the response.
func (m *Modifier) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	if cached, ok := ctx.Get(cachedResponseCtxKey); ok {
		*res = *cached.(*http.Response)
		return nil
	}

	if !m.Update {
		return nil
	}

	key, err := getCacheKey(res.Request)
	if err != nil {
		return fmt.Errorf("cache.Modifier.ModifyResponse(): getCacheKey(): %v", err)
	}

	var buf, body bytes.Buffer
	res.Body = ioutil.NopCloser(io.TeeReader(res.Body, &body))
	defer func() { res.Body = ioutil.NopCloser(&body) }()
	if err := res.Write(&buf); err != nil {
		return fmt.Errorf("cache.Modifier.ModifyResponse(): res.Write(): %v", err)
	}

	return m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(m.Bucket))
		if err != nil {
			return fmt.Errorf("cache.Modifier.ModifyResponse(): CreateBucketIfNotExists(%q): %v", m.Bucket, err)
		}
		if err := b.Put(key, buf.Bytes()); err != nil {
			return fmt.Errorf("cache.Modifier.ModifyResponse(): bucket.Put(): %v", err)
		}
		return nil
	})
}

// getCacheKey gets a cache key from the request context or generates a new one from the request.
func getCacheKey(req *http.Request) ([]byte, error) {
	// Use custom cache key from context if available.
	ctx := martian.NewContext(req)
	if keyRaw, ok := ctx.Get(CustomKey); ok && keyRaw != nil {
		return keyRaw.([]byte), nil
	}
	key, err := generateCacheKey(req)
	if err != nil {
		return nil, fmt.Errorf("generateCacheKey(): %v", err)
	}
	return key, nil
}

// generateCacheKey is a super basic keygen that just hashes the request method and URL.
func generateCacheKey(req *http.Request) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(req.Method)
	b.WriteString(" ")
	b.WriteString(req.URL.String())

	hash := sha1.New()
	if _, err := b.WriteTo(hash); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
