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
	"regexp"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

func init() {
	parse.Register("querystring.Filter", filterFromJSON)
}

// Filter runs modifiers iff the request query parameter for name matches value.
type Filter struct {
	reqmod       martian.RequestModifier
	resmod       martian.ResponseModifier
	nameMatcher  *regexp.Regexp
	valueMatcher *regexp.Regexp
}

type filterJSON struct {
	Name     string               `json:"name"`
	Value    string               `json:"value"`
	Modifier json.RawMessage      `json:"modifier"`
	Scope    []parse.ModifierType `json:"scope"`
}

// NewFilter constructs a querystring.Filter that filters modifiers based on
// query parameters.
func NewFilter(nameMatcher, valueMatcher *regexp.Regexp) (*Filter, error) {
	return &Filter{
		nameMatcher:  nameMatcher,
		valueMatcher: valueMatcher,
	}, nil
}

// SetRequestModifier sets the request modifier for filter.
func (f *Filter) SetRequestModifier(reqmod martian.RequestModifier) {
	f.reqmod = reqmod
}

// SetResponseModifier sets the response modifier for filter.
func (f *Filter) SetResponseModifier(resmod martian.ResponseModifier) {
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

	nameMatcher, err := regexp.Compile(msg.Name)
	if err != nil {
		return nil, err
	}

	valueMatcher, err := regexp.Compile(msg.Value)
	if err != nil {
		return nil, err
	}

	filter, err := NewFilter(nameMatcher, valueMatcher)
	if err != nil {
		return nil, err
	}

	r, err := parse.FromJSON(msg.Modifier)
	if err != nil {
		return nil, err
	}

	reqmod := r.RequestModifier()
	filter.SetRequestModifier(reqmod)

	resmod := r.ResponseModifier()
	filter.SetResponseModifier(resmod)

	return parse.NewResult(filter, msg.Scope)
}

// ModifyRequest applies the contained request modifier if the filter name and values
// match the request query parameters. In the case of a nil valueMatcher, the modifier is
// only applied when the nameMatcher matches.
func (f *Filter) ModifyRequest(ctx *martian.Context, req *http.Request) error {
	if f.matches(req) && f.reqmod != nil {
		return f.reqmod.ModifyRequest(ctx, req)
	}

	return nil
}

// ModifyResponse applies the contained response modifier if the filter name and values match the
// request query parameters.
func (f *Filter) ModifyResponse(ctx *martian.Context, res *http.Response) error {
	if f.matches(res.Request) && f.resmod != nil {
		return f.resmod.ModifyResponse(ctx, res)
	}

	return nil
}

func (f *Filter) matches(req *http.Request) bool {
	var matched bool

	for name, values := range req.URL.Query() {
		if f.nameMatcher.MatchString(name) {
			matched = true
			if f.valueMatcher != nil {
				matched = false
				for _, value := range values {
					if f.valueMatcher.MatchString(value) {
						matched = true
						break
					}
				}
			}
		}
	}

	return matched
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
