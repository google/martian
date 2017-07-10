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
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/google/martian"
	"github.com/google/martian/parse"
)

func init() {
	parse.Register("cache.Modifier", modifierFromJSON)
}

type modifier struct {
	db     *bolt.DB
	bucket string
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
	return nil
}

// ModifyResponse
func (m *modifier) ModifyResponse(res *http.Response) error {
	return nil
}

// NewModifier returns a modifier that
func NewModifier() martian.RequestResponseModifier {
	return nil
}

// modifierFromJSON takes a JSON message as a byte slice and returns a
// cache.Modifier and an error.
//
// Example JSON Configuration message:
// {
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewModifier(), msg.Scope)
}
