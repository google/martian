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
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/google/martian/martiantest"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/verify"

	// Import to register header.Modifier with JSON parser.
	_ "github.com/google/martian/header"
)

func TestFilterWithQueryParamNameAndNoValue(t *testing.T) {
	m, err := regexp.Compile("name")
	if err != nil {
		t.Fatalf("regexp.Compile(%q): got %v, want no error", "name", err)
	}

	filter, err := NewFilter(m, nil)
	if err != nil {
		t.Fatalf("NewFilter(): got %v, want no error", err)
	}

	tm := martiantest.NewModifier()
	filter.SetRequestModifier(tm)

	req, err := http.NewRequest("GET", "http://martian.local?name", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := filter.ModifyRequest(req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}
	if !tm.RequestModified() {
		t.Error("tm.RequestModified(): got false, want true")
	}
	tm.Reset()

	req, err = http.NewRequest("GET", "http://martian.local?test", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := filter.ModifyRequest(req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}
	if tm.RequestModified() {
		t.Error("tm.RequestModified(): got true, want false")
	}
}

func TestFilterWithQueryStringNameAndValue(t *testing.T) {
	nm, err := regexp.Compile("name")
	if err != nil {
		t.Fatalf("regexp.Compile(%q): got %v, want no error", "name", err)
	}
	vm, err := regexp.Compile("value")
	if err != nil {
		t.Fatalf("regexp.Compile(%q): got %v, want no error", "value", err)
	}

	filter, err := NewFilter(nm, vm)
	if err != nil {
		t.Fatalf("NewFilter(): got %v, want no error", err)
	}

	tm := martiantest.NewModifier()
	filter.SetRequestModifier(tm)

	v := url.Values{}
	v.Add("nomatch", "value")
	req, err := http.NewRequest("POST", "http://martian.local?name=value", strings.NewReader(v.Encode()))
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err := filter.ModifyRequest(req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}
	if !tm.RequestModified() {
		t.Error("tm.RequestModified(): got false, want true")
	}
	tm.Reset()

	v = url.Values{}
	req, err = http.NewRequest("POST", "http://martian.local", strings.NewReader(v.Encode()))
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := filter.ModifyRequest(req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}
	if tm.RequestModified() {
		t.Error("tm.RequestModified(): got true, want false")
	}
}

func TestFilterWithQueryStringNameAndNilValueMatcher(t *testing.T) {
	nm, err := regexp.Compile("name")
	if err != nil {
		t.Fatalf("regexp.Compile(%q): got %v, want no error", "name", err)
	}

	filter, err := NewFilter(nm, nil)
	if err != nil {
		t.Fatalf("NewFilter(): got %v, want no error", err)
	}

	tm := martiantest.NewModifier()
	filter.SetRequestModifier(tm)

	req, err := http.NewRequest("GET", "http://martian.local?name=value", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := filter.ModifyRequest(req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}
	if !tm.RequestModified() {
		t.Error("tm.RequestModified(): got false, want true")
	}
	tm.Reset()

	req, err = http.NewRequest("GET", "http://martian.local", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := filter.ModifyRequest(req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}
	if tm.RequestModified() {
		t.Error("tm.RequestModified(): got true, want false")
	}
}

func TestFilterFromJSON(t *testing.T) {
	msg := []byte(`{
		"querystring.Filter": {
      "scope": ["request", "response"],
      "name": "param",
      "value": "true",
      "modifier": {
        "header.Modifier": {
          "scope": ["request", "response"],
          "name": "Mod-Run",
          "value": "true"
        }
      }
    }
	}`)

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got %v, want no error", err)
	}

	reqmod := r.RequestModifier()
	if reqmod == nil {
		t.Fatal("reqmod: got nil, want not nil")
	}

	req, err := http.NewRequest("GET", "https://martian.test?param=true", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("reqmod.ModifyRequest(): got %v, want no error", err)
	}

	if got, want := req.Header.Get("Mod-Run"), "true"; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", "Mod-Run", got, want)
	}

	resmod := r.ResponseModifier()
	if resmod == nil {
		t.Fatalf("resmod: got nil, want not nil")
	}

	res := proxyutil.NewResponse(200, nil, req)
	if err := resmod.ModifyResponse(res); err != nil {
		t.Fatalf("resmod.ModifyResponse(): got %v, want no error", err)
	}

	if got, want := res.Header.Get("Mod-Run"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Mod-Run", got, want)
	}
}

func TestPassThroughVerifyRequests(t *testing.T) {
	f, err := NewFilter(nil, nil)
	if err != nil {
		t.Fatalf("NewFilter(): got %v, want no error", err)
	}

	if err := f.VerifyRequests(); err != nil {
		t.Fatalf("VerifyRequest(): got %v, want no error", err)
	}

	tv := &verify.TestVerifier{
		RequestError: errors.New("verify request failure"),
	}

	f.SetRequestModifier(tv)

	if got, want := f.VerifyRequests(), tv.RequestError; got != want {
		t.Fatalf("VerifyRequests(): got %v, want %v", got, want)
	}
}

func TestPassThroughVerifyResponses(t *testing.T) {
	f, err := NewFilter(nil, nil)
	if err != nil {
		t.Fatalf("NewFilter(): got %v, want no error", err)
	}

	if err := f.VerifyResponses(); err != nil {
		t.Fatalf("VerifyResponses(): got %v, want no error", err)
	}

	tv := &verify.TestVerifier{
		ResponseError: errors.New("verify response failure"),
	}

	f.SetResponseModifier(tv)

	if got, want := f.VerifyResponses(), tv.ResponseError; got != want {
		t.Fatalf("VerifyResponses(): got %v, want %v", got, want)
	}
}

func TestResets(t *testing.T) {
	f, err := NewFilter(nil, nil)
	if err != nil {
		t.Fatalf("NewFilter(): got %v, want no error", err)
	}

	tv := &verify.TestVerifier{
		ResponseError: errors.New("verify response failure"),
	}
	f.SetResponseModifier(tv)

	tv = &verify.TestVerifier{
		RequestError: errors.New("verify request failure"),
	}
	f.SetRequestModifier(tv)

	if err := f.VerifyRequests(); err == nil {
		t.Fatal("VerifyRequests(): got nil, want error")
	}
	if err := f.VerifyResponses(); err == nil {
		t.Fatal("VerifyResponses(): got nil, want error")
	}

	f.ResetRequestVerifications()
	f.ResetResponseVerifications()

	if err := f.VerifyRequests(); err != nil {
		t.Errorf("VerifyRequests(): got %v, want no error", err)
	}
	if err := f.VerifyResponses(); err != nil {
		t.Errorf("VerifyResponses(): got %v, want no error", err)
	}
}
