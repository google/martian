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
	"net/http"
	"net/url"
)

// Matcher is a conditional evaluator of request urls to be used in
// filters that take conditionals.
type Matcher struct {
	url *url.URL
}

// NewMatcher builds a new url matcher.
func NewMatcher(url *url.URL) *Matcher {
	return &Matcher{
		url: url,
	}
}

func (m *Matcher) MatchRequest(req *http.Request) bool {
	return m.matches(req.URL)
}

func (m *Matcher) MatchResponse(res *http.Response) bool {
	return m.matches(res.Request.URL)
}

// matches forces all non-empty URL segments to match or it returns false.
func (m *Matcher) matches(u *url.URL) bool {
	switch {
	case m.url.Scheme != "" && m.url.Scheme != u.Scheme:
		return false
	case m.url.Host != "" && !MatchHost(u.Host, m.url.Host):
		return false
	case m.url.Path != "" && m.url.Path != u.Path:
		return false
	case m.url.RawQuery != "" && m.url.RawQuery != u.RawQuery:
		return false
	case m.url.Fragment != "" && m.url.Fragment != u.Fragment:
		return false
	}

	return true
}
