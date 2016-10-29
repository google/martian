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
	"testing"

	"github.com/google/martian/martiantest"
	"github.com/google/martian/proxyutil"
)

func TestRequestWhenTrueCondition(t *testing.T) {
	filter := New()

	tmc := NewTestMatcher()
	tmc.RequestEvaluatesTo(true)
	filter.SetRequestCondition(tmc)

	tmod := martiantest.NewModifier()
	filter.RequestWhenTrue(tmod)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := filter.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got, want := tmod.RequestModified(), true; got != want {
		t.Errorf("tmod.RequestModified(): got %t, want %t", got, want)
	}
}

func TestRequestWhenFalseCondition(t *testing.T) {
	filter := New()

	tmc := NewTestMatcher()
	tmc.RequestEvaluatesTo(false)
	filter.SetRequestCondition(tmc)

	tmod := martiantest.NewModifier()
	filter.RequestWhenFalse(tmod)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := filter.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got, want := tmod.RequestModified(), true; got != want {
		t.Errorf("tmod.RequestModified(): got %t, want %t", got, want)
	}
}

func TestResponseWhenTrueCondition(t *testing.T) {
	filter := New()

	tmc := NewTestMatcher()
	tmc.ResponseEvaluatesTo(true)
	filter.SetResponseCondition(tmc)

	tmod := martiantest.NewModifier()
	filter.ResponseWhenTrue(tmod)

	res := proxyutil.NewResponse(200, nil, nil)

	if err := filter.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	if got, want := tmod.ResponseModified(), true; got != want {
		t.Errorf("tmod.ResponseModified(): got %t, want %t", got, want)
	}
}

func TestResponseWhenFalseCondition(t *testing.T) {
	filter := New()

	tmc := NewTestMatcher()
	tmc.ResponseEvaluatesTo(false)
	filter.SetResponseCondition(tmc)

	tmod := martiantest.NewModifier()
	filter.ResponseWhenFalse(tmod)

	res := proxyutil.NewResponse(200, nil, nil)

	if err := filter.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	if got, want := tmod.ResponseModified(), true; got != want {
		t.Errorf("tmod.ResponseModified(): got %t, want %t", got, want)
	}
}
