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
	"regexp"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

func init() {
	parse.Register("url.RegexFilter", regexFilterFromJSON)
}

// URLRegexFilter runs modifiers iff the request URL matches the regex. This is not to be confused with
// url.Filter that does string matching on URL segments.
type URLRegexFilter struct {
	reqmod  martian.RequestModifier
	resmod  martian.ResponseModifier
	matcher *regexp.Regexp
}

type regexFilterJSON struct {
	Regex    string               `json:"regex"`
	Modifier json.RawMessage      `json:"modifier"`
	Scope    []parse.ModifierType `json:"scope"`
}

// NewRegexFilter constructs a filter that matches on regular expressions.
func NewRegexFilter(r *regexp.Regexp) *URLRegexFilter {
	return &URLRegexFilter{
		matcher: r,
	}
}

// SetRequestModifier sets the request modifier.
func (f *URLRegexFilter) SetRequestModifier(reqmod martian.RequestModifier) {
	f.reqmod = reqmod
}

// SetResponseModifier sets the response modifier.
func (f *URLRegexFilter) SetResponseModifier(resmod martian.ResponseModifier) {
	f.resmod = resmod
}

// ModifyRequest runs the modifier if the URL matches the provided matcher.
func (f *URLRegexFilter) ModifyRequest(req *http.Request) error {
	if f.reqmod != nil && f.matches(req.URL) {
		return f.reqmod.ModifyRequest(req)
	}

	return nil
}

// ModifyResponse runs the modifier if the request URL matches the provided matcher.
func (f *URLRegexFilter) ModifyResponse(res *http.Response) error {
	if f.resmod != nil && f.matches(res.Request.URL) {
		return f.resmod.ModifyResponse(res)
	}

	return nil
}

// matches applies the regex to the URL.
func (f *URLRegexFilter) matches(u *url.URL) bool {
	return f.matcher.MatchString(u.String())
}

// regexFilterFromJSON takes a JSON message as a byte slice and returns a
// parse.Result that contains a URLRegexFilter and a scope. The regex syntax is RE2
// as described at https://golang.org/s/re2syntax.
//
// Example JSON configuration message:
// {
//   "scope": ["request", "response"],
//   "regex": ".*www.example.com.*"
// }
func regexFilterFromJSON(b []byte) (*parse.Result, error) {
	msg := &regexFilterJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	matcher, err := regexp.Compile(msg.Regex)
	if err != nil {
		return nil, err
	}

	filter := NewRegexFilter(matcher)

	r, err := parse.FromJSON(msg.Modifier)
	if err != nil {
		return nil, err
	}

	reqmod := r.RequestModifier()
	if reqmod != nil {
		filter.SetRequestModifier(reqmod)
	}

	resmod := r.ResponseModifier()
	if resmod != nil {
		filter.SetResponseModifier(resmod)
	}

	return parse.NewResult(filter, msg.Scope)
}