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

// Package header provides utilities for modifying, filtering, and
// verifying headers in martian.Proxy.
package header

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/verify"
)

const (
	missingHeaderFormat = "got no header, want %s header"
	missingValueFormat  = "got %s with value %%s, want value %%s"
)

// Verifier is a header verifier.
type Verifier struct {
	name, value string
}

type verifierJSON struct {
	Name  string               `json:"name"`
	Value string               `json:"value"`
	Scope []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("header.Verifier", verifierFromJSON)
}

// NewVerifier creates a new header verifier for the given name and value.
func NewVerifier(name, value string) *Verifier {
	return &Verifier{
		name:  name,
		value: value,
	}
}

// ModifyRequest verifies that the header for name is present in all modified
// requests. If value is non-empty the value must be present in at least one
// header for name. An error will be added for every unmatched request.
func (v *Verifier) ModifyRequest(req *http.Request) error {
	h := proxyutil.RequestHeader(req)
	ctx := martian.NewContext(req)

	err := verify.NewError("header.Verifier").Request(req)

	return v.verify(ctx, h, err)
}

// ModifyResponse verifies that the header for name is present in all modified
// responses. If value is non-empty the value must be present in at least one
// header for name. An error will be added for every unmatched response.
func (v *Verifier) ModifyResponse(res *http.Response) error {
	h := proxyutil.ResponseHeader(res)
	ctx := martian.NewContext(res.Request)

	err := verify.NewError("header.Verifier").Response(res)

	return v.verify(ctx, h, err)
}

func (v *Verifier) verify(ctx *martian.Context, h *proxyutil.Header, eb *verify.ErrorBuilder) error {
	vs, ok := h.All(v.name)
	if !ok {
		eb.Expected(v.name).Format("got no header, want %s header")
		verify.Verify(ctx, eb)

		return nil
	}

	for _, value := range vs {
		switch v.value {
		case "", value:
			return nil
		}
	}

	eb.Actual(strings.Join(vs, ", ")).
		Expected(v.value).
		Format(fmt.Sprintf("got %s with value %%s, want value %%s", v.name))

	verify.Verify(ctx, eb)

	return nil
}

// verifierFromJSON builds a header.Verifier from JSON.
//
// Example JSON:
// {
//   "name": "header.Verifier",
//   "scope": ["request", "result"],
//   "modifier": {
//     "name": "Martian-Testing",
//     "value": "true"
//   }
// }
func verifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &verifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewVerifier(msg.Name, msg.Value), msg.Scope)
}
