// Copyright 2015 Google Inc. All rights reserved.
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

// Package body allows for the replacement of message body on responses.
package body

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/google/martian/parse"
)

func init() {
	parse.Register("body.Modifier", modifierFromJSON)
}

// Modifier substitutes the body on an HTTP response.
type Modifier struct {
	contentType string
	body        []byte
}

type modifierJSON struct {
	ContentType string               `json:"contentType"`
	Path        string               `json:"path"` // Path is a path local to the proxy.
	Body        []byte               `json:"body"` // Body is expected to be a Base64 encoded string.
	Scope       []parse.ModifierType `json:"scope"`
}

// NewModifier constructs and returns a body.Modifier.
func NewModifier(b []byte, contentType string) *Modifier {
	return &Modifier{
		contentType: contentType,
		body:        b,
	}
}

// NewModifierFromFile returns a body.Modifier that substitutes the body on an
// HTTP response with bytes read from a file local to the proxy.
func NewModifierFromFile(path string, contentType string) (*Modifier, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &Modifier{
		contentType: contentType,
		body:        b,
	}, nil
}

// modifierFromJSON takes a JSON message as a byte slice and returns a
// body.Modifier and an error.
//
// Example JSON Configuration message providing Base64 encoded body:
// {
//   "scope": ["request", "response"],
//   "contentType": "text/plain",
//   "body": "c29tZSBkYXRhIHdpdGggACBhbmQg77u/" // Base64 encoded body
// }
//
// Example JSON Configuration message providing a local file:
// {
//   "scope": ["request", "response"],
//   "contentType": "text/plain",
//   "path": "some/local/path/to/a/file.json/" // Path local to the proxy
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	if msg.Body != nil && msg.Path != "" {
		return nil, errors.New("body.modifierFromJSON: Body and Path both supplied. These fields are mutually exclusive")
	}

	if msg.Body != nil {
		mod := NewModifier(msg.Body, msg.ContentType)

		return parse.NewResult(mod, msg.Scope)
	}

	if msg.Path != "" {
		mod, err := NewModifierFromFile(msg.Path, msg.ContentType)
		if err != nil {
			return nil, err
		}

		return parse.NewResult(mod, msg.Scope)
	}

	return nil, errors.New("body.modifierFromJSON: Neither Body nor Path supplied.")
}

// ModifyRequest sets the Content-Type header and overrides the request body.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	req.Body.Close()

	req.Header.Set("Content-Type", m.contentType)

	// Reset the Content-Encoding since we know that the new body isn't encoded.
	req.Header.Del("Content-Encoding")

	req.ContentLength = int64(len(m.body))
	req.Body = ioutil.NopCloser(bytes.NewReader(m.body))

	return nil
}

// ModifyResponse sets the Content-Type header and overrides the response body.
func (m *Modifier) ModifyResponse(res *http.Response) error {
	// Replace the existing body, close it first.
	res.Body.Close()

	res.Header.Set("Content-Type", m.contentType)

	// Reset the Content-Encoding since we know that the new body isn't encoded.
	res.Header.Del("Content-Encoding")

	res.ContentLength = int64(len(m.body))
	res.Body = ioutil.NopCloser(bytes.NewReader(m.body))

	return nil
}
