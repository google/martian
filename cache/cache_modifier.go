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
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/martian"
	"github.com/google/martian/parse"
)

const (
	// CustomKey is the context key for setting a custom cache key for a request.
	CustomKey = "cache.CustomKey"

	// cachedResponseCtxKey is the context key for storing the cached response for a request.
	cachedResponseCtxKey = "cache.Response"
)

func init() {
	parse.Register("cache.Modifier", modifierFromJSON)
}

type modifier struct {
	db       *bolt.DB
	bucket   string
	update   bool
	replay   bool
	hermetic bool
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
// `filepath` is the filepath to the boltdb file containing cached responses.
// `bucket` is the bucket name of the boltdb to use.
// If `update` is true, the database will be updated with any live responses, e.g. on cache miss or when not replaying.
// If `replay` is true, the modifier will replay responses from its cache.
// If `hermetic` is true, the modifier will return error if it cannot replay a cached response, e.g. on cache miss or not replaying.
// Argument combinations that don't make sense will return error, e.g. replay=false and hermetic=true.
func NewModifier(filepath, bucket string, update, replay, hermetic bool) (martian.RequestResponseModifier, error) {
	if !replay && hermetic {
		return nil, fmt.Errorf("cache.Modifier: cannot use hermetic mode if not replaying")
	}

	opt := &bolt.Options{
		Timeout:  10 * time.Second,
		ReadOnly: !update,
	}
	log.Printf("cache.Modifier: opening boltdb file %q", filepath)
	db, err := bolt.Open(filepath, 0600, opt)
	if err != nil {
		return nil, fmt.Errorf("cache.Modifier: bolt.Open(%q): %v", filepath, err)
	}
	// TODO: Figure out how to close the db after use.

	if bucket != "" && update {
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("CreateBucketIfNotExists(%q): %s", bucket, err)
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("cache.Modifier: db.Update(): %v", err)
		}
	}

	return &modifier{
		db:       db,
		bucket:   bucket,
		update:   update,
		replay:   replay,
		hermetic: hermetic,
	}, nil
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
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, fmt.Errorf("cache.Modifier: json.Unmarshal(): %v", err)
	}

	mod, err := NewModifier(msg.File, msg.Bucket, msg.Update, msg.Replay, msg.Hermetic)
	if err != nil {
		return nil, err
	}
	return parse.NewResult(mod, msg.Scope)
}

// ModifyRequest modifies the http.Request.
func (m *modifier) ModifyRequest(req *http.Request) error {
	if !m.replay {
		return nil
	}

	key, err := getCacheKey(req)
	if err != nil {
		return fmt.Errorf("cache.Modifier: getCacheKey(): %v", err)
	}

	if err := m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(m.bucket))
		cached := b.Get(key)
		if cached != nil {
			res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(cached)), req)
			if err != nil {
				return fmt.Errorf("http.ReadResponse(): %v", err)
			}
			ctx := martian.NewContext(req)
			ctx.SkipRoundTrip()
			ctx.Set(cachedResponseCtxKey, res)
			return nil
		} else if m.hermetic {
			return fmt.Errorf("in hermetic mode and no cached response found")
		}
		return nil
	}); err != nil {
		return fmt.Errorf("cache.Modifier: %v", err)
	}
	return nil
}

// ModifyResponse modifies the http.Response.
func (m *modifier) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	cached, ok := ctx.Get(cachedResponseCtxKey)
	if ok {
		*res = *cached.(*http.Response)
	} else if m.update {
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
			b := tx.Bucket([]byte(m.bucket))
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
