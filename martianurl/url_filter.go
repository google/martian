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

package martianurl

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

func init() {
	parse.Register("url.Filter", filterFromJSON)
}

// Filter runs modifiers iff the request URL matches all of the segments in url.
type Filter struct {
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier
	url    *url.URL
}

type filterJSON struct {
	Scheme   string               `json:"scheme"`
	Host     string               `json:"host"`
	Path     string               `json:"path"`
	Query    string               `json:"query"`
	Modifier json.RawMessage      `json:"modifier"`
	Scope    []parse.ModifierType `json:"scope"`
}

// NewFilter constructs a filter that applies the modifer when the
// request URL matches all of the provided URL segments.
func NewFilter(u *url.URL) *Filter {
	return &Filter{
		url: u,
	}
}

// SetRequestModifier sets the request modifier.
func (f *Filter) SetRequestModifier(reqmod martian.RequestModifier) {
	f.reqmod = reqmod
}

// SetResponseModifier sets the response modifier.
func (f *Filter) SetResponseModifier(resmod martian.ResponseModifier) {
	f.resmod = resmod
}

// ModifyRequest runs the modifier if the URL matches all provided matchers.
func (f *Filter) ModifyRequest(req *http.Request) error {
	if f.reqmod != nil && f.matches(req.URL) {
		return f.reqmod.ModifyRequest(req)
	}

	return nil
}

// ModifyResponse runs the modifier if the request URL matches urlMatcher.
func (f *Filter) ModifyResponse(res *http.Response) error {
	if f.resmod != nil && f.matches(res.Request.URL) {
		return f.resmod.ModifyResponse(res)
	}

	return nil
}

// filterFromJSON takes a JSON message as a byte slice and returns a
// parse.Result that contians a URLFilter and a bitmask that represents the
// type of modifier.
//
// Example JSON configuration message:
// {
//   "scheme": "https",
//   "host": "example.com",
//   "path": "/foo/bar",
//   "rawQuery": "q=value",
//   "scope": ["request", "response"],
//   "modifier": { ... }
// }
func filterFromJSON(b []byte) (*parse.Result, error) {
	msg := &filterJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	filter := NewFilter(&url.URL{
		Scheme:   msg.Scheme,
		Host:     msg.Host,
		Path:     msg.Path,
		RawQuery: msg.Query,
	})

	r, err := parse.FromJSON(msg.Modifier)
	if err != nil {
		return nil, err
	}

	reqmod := r.RequestModifier()
	if err != nil {
		return nil, err
	}
	if reqmod != nil {
		filter.SetRequestModifier(reqmod)
	}

	resmod := r.ResponseModifier()
	if resmod != nil {
		filter.SetResponseModifier(resmod)
	}

	return parse.NewResult(filter, msg.Scope)
}

// matches forces all non-empty URL segments to match or it returns false.
func (f *Filter) matches(u *url.URL) bool {
	switch {
	case f.url.Scheme != "" && f.url.Scheme != u.Scheme:
		return false
	case f.url.Host != "" && f.url.Host != u.Host:
		return false
	case f.url.Path != "" && f.url.Path != u.Path:
		return false
	case f.url.RawQuery != "" && f.url.RawQuery != u.RawQuery:
		return false
	case f.url.Fragment != "" && f.url.Fragment != u.Fragment:
		return false
	}

	return true
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
