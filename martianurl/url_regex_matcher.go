// Copyright 2017 Google Inc. All rights reserved.
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
	"net/http"
	"net/url"
	"regexp"
)

// Matcher is a conditional evaluator of request urls to be used in
// filters that take conditionals.
type RegexMatcher struct {
	r *regexp.Regexp
}

// NewMatcher builds a new url matcher.
func NewRegexMatcher(r *regexp.Regexp) *RegexMatcher {
	return &RegexMatcher{
		r: r,
	}
}

// MatchRequest retuns true if all non-empty URL segments in m.url match the
// request URL.
func (m *RegexMatcher) MatchRequest(req *http.Request) bool {
	return m.matches(req.URL)
}

// MatchResponse retuns true if all non-empty URL segments in m.url match the
// request URL.
func (m *RegexMatcher) MatchResponse(res *http.Response) bool {
	return m.matches(res.Request.URL)
}

// matches forces all non-empty URL segments to match or it returns false.
func (m *RegexMatcher) matches(u *url.URL) bool {
	return m.r.MatchString(u.String())
}
