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
	"fmt"
	"net/http"

	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
)

func init() {
	parse.Register("header.Modifier", modifierFromJSON)
}

// HeaderSetBehavior defines the behavior of setting a new header that already exists.
type HeaderSetBehavior int

const (
	// ReplaceValues replaces all previous headers with the same canonicalized name.
	ReplaceValues HeaderSetBehavior = iota
	// AppendValues appends the new header without replacing existing ones.
	AppendValues
)

// Modifier adds or replaces headers to requests and responses.
type Modifier struct {
	name, value string
	behavior    HeaderSetBehavior
}

type modifierJSON struct {
	Name     string               `json:"name"`
	Value    string               `json:"value"`
	Behavior string               `json:"behavior"`
	Scope    []parse.ModifierType `json:"scope"`
}

// ModifyRequest sets the header at name with value on the request.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	h := proxyutil.RequestHeader(req)
	if m.behavior == AppendValues {
		return h.Add(m.name, m.value)
	}
	return h.Set(m.name, m.value)
}

// ModifyResponse sets the header at name with value on the response.
func (m *Modifier) ModifyResponse(res *http.Response) error {
	h := proxyutil.ResponseHeader(res)
	if m.behavior == AppendValues {
		return h.Add(m.name, m.value)
	}
	return h.Set(m.name, m.value)
}

// SetBehavior sets ReplaceValues or AppendValues as the behavior of the modifier
// when a header already exists.
func (m *Modifier) SetBehavior(b HeaderSetBehavior) {
	m.behavior = b
}

// NewModifier returns a modifier that will set the header at name with
// the given value for both requests and responses. By default, if the header name
// already exists all values will be overwritten.
func NewModifier(name, value string) *Modifier {
	return &Modifier{
		name:     http.CanonicalHeaderKey(name),
		value:    value,
		behavior: ReplaceValues,
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

	switch msg.Behavior {
	case "replace", "":
		modifier.SetBehavior(ReplaceValues)
	case "append":
		modifier.SetBehavior(AppendValues)
	default:
		return nil, fmt.Errorf("Invalid header modifier behavior %q. Must be either append or replace", msg.Behavior)
	}

	return parse.NewResult(modifier, msg.Scope)
}
