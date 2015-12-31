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

// Package method provides utilities for verifying method type in martian.Proxy.
package method

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

// Verifier is a method verifier.
type Verifier struct {
	method string
}

type verifierJSON struct {
	Method string               `json:"method"`
	Scope  []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("method.Verifier", verifierFromJSON)
}

// NewVerifier returns a new method verifier.
func NewVerifier(method string) (*Verifier, error) {
	if method == "" {
		return nil, fmt.Errorf("method: method cannot be blank")
	}
	return &Verifier{
		method: method,
	}, nil
}

// ModifyRequest verifies that the request method matches the given method in
// all modified requests. An error will be added if the method does not match.
func (v *Verifier) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)

	if v.method != req.Method {
		eb := verify.NewError("method.Verifier").
			Request(req).
			Actual(req.Method).
			Expected(v.method)

		verify.Verify(ctx, eb)
	}

	return nil
}

// verifierFromJSON builds a method.Verifier from JSON.
//
// Example JSON:
// {
//   "method.Verifier": {
//     "scope": ["request"],
//     "method": "POST"
//   }
// }
func verifierFromJSON(b []byte) (*parse.Result, error) {

	msg := &verifierJSON{}

	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}
	v, err := NewVerifier(msg.Method)
	if err != nil {
		return nil, err
	}
	return parse.NewResult(v, msg.Scope)
}
