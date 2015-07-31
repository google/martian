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

package header

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

func init() {
	parse.Register("header.Modifier", modifierFromJSON)
}

type modifier struct {
	name, value string
}

type modifierJSON struct {
	Name  string               `json:"name"`
	Value string               `json:"value"`
	Scope []parse.ModifierType `json:"scope"`
}

// ModifyRequest sets the header at name with value on the request.
func (m *modifier) ModifyRequest(req *http.Request) error {
	// Host is treated differently by the http package and the header is
	// explicitly ignored.
	if m.name == "Host" {
		req.Host = m.value
	} else {
		req.Header.Set(m.name, m.value)
	}

	return nil
}

// ModifyResponse sets the header at name with value on the response.
func (m *modifier) ModifyResponse(res *http.Response) error {
	res.Header.Set(m.name, m.value)

	return nil
}

// NewModifier returns a modifier that will set the header at name with
// the given value for both requests and responses. If the header name already
// exists all values will be overwritten.
func NewModifier(name, value string) martian.RequestResponseModifier {
	return &modifier{
		name:  http.CanonicalHeaderKey(name),
		value: value,
	}
}

// modifierFromJSON takes a JSON message as a byte slice and returns
// a headerModifier and an error.
//
// Example JSON configuration message:
// {
//  "scope": ["request", "result"],
//  "name": "X-Martian",
//  "value": "true"
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	modifier := NewModifier(msg.Name, msg.Value)

	return parse.NewResult(modifier, msg.Scope)
}
