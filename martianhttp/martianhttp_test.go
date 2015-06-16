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

package martianhttp

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/verify"

	_ "github.com/google/martian/header"
)

func TestModifyRequestNoModifier(t *testing.T) {
	m := NewModifier()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := m.ModifyRequest(martian.NewContext(), req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}
}

func TestModifyRequest(t *testing.T) {
	m := NewModifier()

	var modRun bool
	m.reqmod = martian.RequestModifierFunc(
		func(*martian.Context, *http.Request) error {
			modRun = true
			return nil
		})

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := m.ModifyRequest(martian.NewContext(), req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if !modRun {
		t.Error("modRun: got false, want true")
	}
}

func TestModifyResponse(t *testing.T) {
	m := NewModifier()

	var modRun bool
	m.resmod = martian.ResponseModifierFunc(
		func(*martian.Context, *http.Response) error {
			modRun = true
			return nil
		})

	res := proxyutil.NewResponse(200, nil, nil)
	if err := m.ModifyResponse(martian.NewContext(), res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if !modRun {
		t.Error("modRun: got false, want true")
	}
}

func TestModifyResponseNoModifier(t *testing.T) {
	m := NewModifier()
	res := proxyutil.NewResponse(200, nil, nil)

	if err := m.ModifyResponse(martian.NewContext(), res); err != nil {
		t.Errorf("ModifyResponse(): got %v, want no error", err)
	}
}

func TestVerifyRequestsNoVerifier(t *testing.T) {
	m := NewModifier()

	if err := m.VerifyRequests(); err != nil {
		t.Errorf("VerifyRequests(): got %v, want no error", err)
	}
}

func TestVerifyRequests(t *testing.T) {
	m := NewModifier()
	verr := fmt.Errorf("request verification failure")

	m.reqmod = &verify.TestVerifier{
		RequestError: verr,
	}

	if err := m.VerifyRequests(); err != verr {
		t.Errorf("VerifyRequests(): got %v, want %v", err, verr)
	}
}

func TestVerifyResponsesNoVerifier(t *testing.T) {
	m := NewModifier()

	if err := m.VerifyResponses(); err != nil {
		t.Errorf("VerifyResponses(): got %v, want no error", err)
	}
}

func TestVerifyResponses(t *testing.T) {
	m := NewModifier()
	verr := fmt.Errorf("response verification failure")

	m.resmod = &verify.TestVerifier{
		ResponseError: verr,
	}

	if err := m.VerifyResponses(); err != verr {
		t.Errorf("VerifyResponses(): got %v, want %v", err, verr)
	}
}

func TestServeHTTPInvalidMethod(t *testing.T) {
	m := NewModifier()

	req, err := http.NewRequest("GET", "/martian/modifiers", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q, ...): got %v, want no error", "GET", err)
	}
	rw := httptest.NewRecorder()

	m.ServeHTTP(rw, req)
	if got, want := rw.Code, 405; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}
	if got, want := rw.Header().Get("Allow"), "POST"; got != want {
		t.Errorf("rw.Header().Get(%q): got %q, want %q", "Allow", got, want)
	}
}

func TestServeHTTPInvalidJSON(t *testing.T) {
	m := NewModifier()

	req, err := http.NewRequest("POST", "/martian/modifiers", bytes.NewBuffer([]byte("not-json")))
	if err != nil {
		t.Fatalf("http.NewRequest(%q, %q, ...): got %v, want no error", "POST", "/martian/modifiers", err)
	}
	rw := httptest.NewRecorder()

	m.ServeHTTP(rw, req)
	if got, want := rw.Code, 400; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}
}

func TestServeHTTP(t *testing.T) {
	m := NewModifier()

	msg := []byte(`
	{
    "header.Modifier": {
      "scope": ["request", "response"],
			"name": "Martian-Test",
			"value": "true"
		}
	}`)

	req, err := http.NewRequest("POST", "/martian/modifiers?id=id", bytes.NewBuffer(msg))
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw := httptest.NewRecorder()

	m.ServeHTTP(rw, req)
	if got, want := rw.Code, 200; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}

	req, err = http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := m.ModifyRequest(martian.NewContext(), req); err != nil {
		t.Fatalf("m.ModifyRequest(): got %v, want no error", err)
	}
	if got, want := req.Header.Get("Martian-Test"), "true"; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", "Martian-Test", got, want)
	}

	res := proxyutil.NewResponse(200, nil, req)
	if err := m.ModifyResponse(martian.NewContext(), res); err != nil {
		t.Fatalf("m.ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.Header.Get("Martian-Test"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Martian-Test", got, want)
	}
}
