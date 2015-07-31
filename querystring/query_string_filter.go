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
	"net/http"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

var noop = martian.Noop("querystring.Filter")

func init() {
	parse.Register("querystring.Filter", filterFromJSON)
}

// Filter runs modifiers iff the request query parameter for name matches value.
type Filter struct {
	name   string
	value  string
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier
}

type filterJSON struct {
	Name     string               `json:"name"`
	Value    string               `json:"value"`
	Modifier json.RawMessage      `json:"modifier"`
	Scope    []parse.ModifierType `json:"scope"`
}

// NewFilter builds a querystring.Filter that filters on name and optionally
// value.
func NewFilter(name, value string) *Filter {
	return &Filter{
		name:   name,
		value:  value,
		reqmod: noop,
		resmod: noop,
	}
}

// SetRequestModifier sets the request modifier for filter.
func (f *Filter) SetRequestModifier(reqmod martian.RequestModifier) {
	if reqmod == nil {
		reqmod = noop
	}

	f.reqmod = reqmod
}

// SetResponseModifier sets the response modifier for filter.
func (f *Filter) SetResponseModifier(resmod martian.ResponseModifier) {
	if resmod == nil {
		resmod = noop
	}

	f.resmod = resmod
}

// filterFromJSON takes a JSON message and returns a querystring.Filter.
//
// Example JSON:
// {
//   "name": "param",
//   "value": "example",
//   "scope": ["request", "response"],
//   "modifier": { ... }
// }
func filterFromJSON(b []byte) (*parse.Result, error) {
	msg := &filterJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	f := NewFilter(msg.Name, msg.Value)

	r, err := parse.FromJSON(msg.Modifier)
	if err != nil {
		return nil, err
	}

	f.SetRequestModifier(r.RequestModifier())
	f.SetResponseModifier(r.ResponseModifier())

	return parse.NewResult(f, msg.Scope)
}

// ModifyRequest applies the request modifier if the filter name and values
// match the request query parameters. In the case of an empty value, the modifier is
// applied whenever any parameter matches name, regardless of its value.
func (f *Filter) ModifyRequest(req *http.Request) error {
	if f.matches(req) && f.reqmod != nil {
		return f.reqmod.ModifyRequest(req)
	}

	return nil
}

// ModifyResponse applies the response modifier if the filter name and values
// match the request query parameters. In the case of an empty value, the modifier is
// applied whenever any parameter matches name, regardless of its value.
func (f *Filter) ModifyResponse(res *http.Response) error {
	if f.matches(res.Request) && f.resmod != nil {
		return f.resmod.ModifyResponse(res)
	}

	return nil
}

func (f *Filter) matches(req *http.Request) bool {
	for n, vs := range req.URL.Query() {
		if f.name == n {
			if f.value == "" {
				return true
			}

			for _, v := range vs {
				if f.value == v {
					return true
				}
			}
		}
	}

	return false
}

// VerifyRequests returns an error containing all the verification errors
// returned by request verifiers.
func (f *Filter) VerifyRequests() error {
	if reqv, ok := f.reqmod.(verify.RequestVerifier); ok {
		return reqv.VerifyRequests()
	}

	return nil
}

// VerifyResponses returns an error containing all the verification errors
// returned by response verifiers.
func (f *Filter) VerifyResponses() error {
	if resv, ok := f.resmod.(verify.ResponseVerifier); ok {
		return resv.VerifyResponses()
	}

	return nil
}

// ResetRequestVerifications resets the state of the contained request verifiers.
func (f *Filter) ResetRequestVerifications() {
	if reqv, ok := f.reqmod.(verify.RequestVerifier); ok {
		reqv.ResetRequestVerifications()
	}
}

// ResetResponseVerifications resets the state of the contained response verifiers.
func (f *Filter) ResetResponseVerifications() {
	if resv, ok := f.resmod.(verify.ResponseVerifier); ok {
		resv.ResetResponseVerifications()
	}
}
