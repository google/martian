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
	db       *bolt.DB
	// Bucket is the name of the database bucket.
	Bucket   string
	// Update determines whether the cache will be updated with new responses.
	Update   bool
	// Replay determines whether the modifier will replay cached responses.
	Replay   bool
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
// The returned modifier will be in non-hermetic replay mode using a default bucket name.
// `filepath` is the filepath to the boltdb file containing cached responses.
func NewModifier(filepath string) (*Modifier, error) {
	log.Infof("cache.Modifier: opening boltdb file %q", filepath)
	db, err := bolt.Open(filepath, 0600, &bolt.Options{
		Timeout: defaultFileOpTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("cache.Modifier: bolt.Open(%q): %v", filepath, err)
	}
	runtime.SetFinalizer(db, func(db *bolt.DB){
		log.Infof("cache.Modifier: closing boltdb file %q", db.Path())
		if err := db.Close(); err != nil {
			log.Errorf("cache.Modifier: db.Close(): %v", err)
		}
		})
	// TODO: Figure out how to close the db after use.

	return &Modifier{
		db:       db,
		Bucket:   defaultBucket,
		Update:   false,
		Replay:   true,
		Hermetic: false,
	}, nil
}

// `bucket` is the bucket name of the boltdb to use. It will be created in the db if it doesn't already exist.
// If `update` is true, the database will be updated with any live responses, e.g. on cache miss or when not replaying.
// If `replay` is true, the modifier will replay responses from its cache.
// If `hermetic` is true, the modifier will return error if it cannot replay a cached response, e.g. on cache miss or not replaying.
// Argument combinations that don't make sense will return error, e.g. replay=false and hermetic=true.

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

	if mod.Update {
		if err := mod.db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(mod.Bucket))
			if err != nil {
				return fmt.Errorf("CreateBucketIfNotExists(%q): %v", mod.Bucket, err)
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("cache.Modifier: db.Update(): %v", err)
		}
	}

	log.Infof("cache.Modifier: created modifier from JSON: %v", mod)
	return parse.NewResult(mod, msg.Scope)
}

// String returns a string representation of the modifier.
func (m *Modifier) String() string {
	return fmt.Sprintf("cache.Modifier{File: %q, Bucket: %q, Update: %t, Replay: %t, Hermetic: %t}",
		m.db.Path(), m.Bucket, m.Update, m.Replay, m.Hermetic)
}

// ModifyRequest modifies the http.Request.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	if !m.Replay {
		if m.Hermetic {
			return errors.New("cache.Modifier: in hermetic mode and not replaying from cache")
		}
		return nil
	}
fmt.Println("getCacheKey()")
	key, err := getCacheKey(req)
	if err != nil {
		return fmt.Errorf("cache.Modifier: getCacheKey(): %v", err)
	}
	fmt.Printf("key=%v\n", key)
fmt.Println("m.db.View()")
	if err := m.db.View(func(tx *bolt.Tx) error {
		fmt.Printf("tx.Bucket(%v)\n", m.Bucket)
		b := tx.Bucket([]byte(m.Bucket))
fmt.Printf("b=%v\n", b)
		fmt.Println("b.Get(key)")
		cached := b.Get(key)
		if cached != nil {
			fmt.Println("ReadResponse()")
			res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(cached)), req)
			if err != nil {
				return fmt.Errorf("http.ReadResponse(): %v", err)
			}
			fmt.Println("NewContext()")
			ctx := martian.NewContext(req)
			ctx.SkipRoundTrip()
			ctx.Set(cachedResponseCtxKey, res)
			return nil
		}
		if m.Hermetic {
			return errors.New("in hermetic mode and no cached response found")
		}
		return nil
	}); err != nil {
		return fmt.Errorf("cache.Modifier: %v", err)
	}
	return nil
}

// ModifyResponse modifies the http.Response.
func (m *Modifier) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	if cached, ok := ctx.Get(cachedResponseCtxKey); ok {
		*res = *cached.(*http.Response)
		return nil
	}
	if m.Update {
		key, err := getCacheKey(res.Request)
		if err != nil {
			return fmt.Errorf("cache.Modifier: getCacheKey(): %v", err)
		}
		var buf bytes.Buffer
		// TODO: Wrap res.Body so res.Write doesn't close it.
		if err := res.Write(&buf); err != nil {
			return fmt.Errorf("cache.Modifier: res.Write(): %v", err)
		}
		if err := m.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(m.Bucket))
			if err := b.Put(key, buf.Bytes()); err != nil {
				return fmt.Errorf("bucket.Put(): %v", err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("cache.Modifier: %v", err)
		}
		r, err := http.ReadResponse(bufio.NewReader(&buf), res.Request)
		if err != nil {
			return fmt.Errorf("cache.Modifier: http.ReadResponse(): %v", err)
		}
		*res = *r
	}
	return nil
}

// getCacheKey gets a cache key from the request context or generates a new one.
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
