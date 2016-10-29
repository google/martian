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

package filter

import (
	"net/http"
)

// ResponseCondition is the interface that describes matchers for response filters
type ResponseCondition interface {
	MatchResponse(*http.Response) bool
}

// RequestCondition is the interface that describes matchers for response filters
type RequestCondition interface {
	MatchRequest(*http.Request) bool
}

// TestMatcher is a stubbed matcher used in tests.
type TestMatcher struct {
	resval bool
	reqval bool
}

// NewTestMatcher returns a pointer to TestMatcher with the return values for
// MatchRequest and MatchResponse intiailized to true.
func NewTestMatcher() *TestMatcher {
	return &TestMatcher{resval: true, reqval: true}
}

// ResponseEvaluatesTo sets the value returned by MatchResponse.
func (tm *TestMatcher) ResponseEvaluatesTo(value bool) {
	tm.resval = value
}

// RequestEvaluatesTo sets the value returned by MatchRequest.
func (tm *TestMatcher) RequestEvaluatesTo(value bool) {
	tm.reqval = value
}

// MatchRequest returns the stubbed value in tm.reqval.
func (tm *TestMatcher) MatchRequest(*http.Request) bool {
	return tm.reqval
}

// MatchResponse returns the stubbed value in tm.resval.
func (tm *TestMatcher) MatchResponse(*http.Response) bool {
	return tm.resval
}
