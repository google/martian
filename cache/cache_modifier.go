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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/martian"
	"github.com/google/martian/log"
	"github.com/google/martian/parse"
)

func init() {
	parse.Register("cache.Modifier", modifierFromJSON)
}

type modifier struct {
	db     *bolt.DB
	bucket string
	replay bool
	update bool
}

type modifierJSON struct {
	File   string               `json:"file"`
	Bucket string               `json:"bucket"`
	Replay bool                 `json:"replay"`
	Update bool                 `json:"update"`
	Scope  []parse.ModifierType `json:"scope"`
}

// ModifyRequest
func (m *modifier) ModifyRequest(req *http.Request) error {
	log.Infof("Modifying request with bucket %s: %v", m.bucket, *req.URL)
	return nil
}

// ModifyResponse
func (m *modifier) ModifyResponse(res *http.Response) error {
	log.Infof("Modifying response with bucket %s: %v", m.bucket, *res.Request.URL)
	return nil
}

// NewModifier returns a modifier that
func NewModifier(filepath, bucket string, replay, update bool) (martian.RequestResponseModifier, error) {
	log.Infof("Making new cache.Modifier to %s", filepath)
	opt := &bolt.Options{
		Timeout:  10 * time.Second,
		ReadOnly: !update,
	}
	log.Infof("cache.Modifier: opening boltdb file %q", filepath)
	db, err := bolt.Open(filepath, 0600, opt)
	if err != nil {
		return nil, fmt.Errorf("bolt.Open(%q): %v", filepath, err)
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
			return nil, err
		}
	}

	mod := &modifier{
		db:     db,
		bucket: bucket,
		replay: replay,
		update: update,
	}
	// runtime.SetFinalizer(m, func(m *modifier) {
	// 	filepath := filepath
	// 	log.Infof("Closing db with file %s", filepath)
	// 	})
	// runtime.SetFinalizer(m.db, func(db *something) {
	// 	log.Infof("Releasing mutex")
	// 	// db.mu.Unlock()
	// 	// log.Infof("Closing db with file %s", *db)
	// 	// fmt.Infof("Closing db with file %s", *db)
	// })
	return mod, nil
}

// modifierFromJSON takes a JSON message as a byte slice and returns a
// cache.Modifier and an error.
//
// Example JSON Configuration message:
// {
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	log.Infof("Modifier fron JSON!")
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	log.Infof("Making new modifier")
	mod, err := NewModifier(msg.File, msg.Bucket, msg.Replay, msg.Update)
	if err != nil {
		return nil, fmt.Errorf("cache.NewModifier: %v", err)
	}
	log.Infof("Made new modifier with bucket %s with file %s", msg.Bucket, msg.File)
	return parse.NewResult(mod, msg.Scope)
}
