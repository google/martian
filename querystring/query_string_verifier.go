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

package querystring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

func init() {
	parse.Register("querystring.Verifier", verifierFromJSON)
}

// Verifier is a verifier for query strings.
type Verifier struct {
	name  string
	value string
}

type verifierJSON struct {
	Name  string               `json:"name"`
	Value string               `json:"value"`
	Scope []parse.ModifierType `json:"scope"`
}

// NewVerifier returns a new query string verifier.
func NewVerifier(name, value string) (*Verifier, error) {
	if name == "" {
		return nil, fmt.Errorf("querystring: parameter name cannot be blank")
	}

	return &Verifier{
		name:  name,
		value: value,
	}, nil
}

// ModifyRequest verifies that the request URL query string parameter matches
// the given name and value in all modified requests. If no value is provided,
// the verifier will only verifier the given name is present. An error will be
// added if the query string parameter is unmatched.
func (v *Verifier) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	eb := verify.NewError("querystring.Verifier").Request(req)

	vs, ok := req.URL.Query()[v.name]
	if !ok {
		eb.Expected(v.name).Format("got no query string parameter, want %s parameter")

		verify.Verify(ctx, eb)

		return nil
	}

	// No value verification required, pass.
	if v.value == "" {
		return nil
	}

	for _, vl := range vs {
		// Value found, pass.
		if v.value == vl {
			return nil
		}
	}

	eb.Actual(strings.Join(vs, ", ")).
		Expected(v.value).
		Format(fmt.Sprintf("%s: got %%s, want to contain %%s", v.name))

	verify.Verify(ctx, eb)

	return nil
}

// verifierFromJSON builds a querystring.Verifier from JSON.
//
// Example JSON:
// {
//   "querystring.Verifier": {
//     "scope": ["request", "response"],
//     "name": "Martian-Testing",
//     "value": "true"
//   }
// }
func verifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &verifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	v, err := NewVerifier(msg.Name, msg.Value)
	if err != nil {
		return nil, err
	}

	return parse.NewResult(v, msg.Scope)
}
