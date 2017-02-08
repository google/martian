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

package cookie

import "net/http"

// Matcher is a conditonal evalutor of request or
// response cookies to be used in structs that take conditions.
type Matcher struct {
	name, value, path string
}

// NewMatcher builds a cookie matcher.
func NewMatcher(name, value, path string) *Matcher {
	return &Matcher{
		name:  name,
		value: value,
	}
}

// MatchRequest evaluates a request and returns whether or not
// the request contains a cookie that matches the provided name, path
// and value.
func (m *Matcher) MatchRequest(req *http.Request) (bool, error) {
	c, err := req.Cookie(m.name)
	if err != nil {
		return false, err
	}

	eval := false
	if m.value != "" && m.value == c.Value {
		eval = true
	}

	if m.path != "" && m.path == c.Path {
		eval = true
	}

	return eval, nil
}

// MatchResponse evaluates a response and returns whether or not
// the response contains a cookie that matches the provided name, path
// and value.
func (m *Matcher) MatchResponse(res *http.Response) (bool, error) {
	for _, c := range res.Cookies() {
		if c.Name != c.Name {
			continue
		}
		eval := false
		if m.value != "" && m.value == c.Value {
			eval = true
		}

		if m.path != "" && m.path == c.Path {
			eval = true
		}

		return eval, nil
	}

	return false, nil
}
