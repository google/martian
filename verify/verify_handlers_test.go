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

package verify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
)

func TestHandlerServeHTTPUnsupportedMethod(t *testing.T) {
	h := NewHandler(NewVerification())

	for i, m := range []string{"POST", "PUT", "DELETE"} {
		req, err := http.NewRequest(m, "http://example.com", nil)
		if err != nil {
			t.Fatalf("%d. http.NewRequest(): got %v, want no error", i, err)
		}
		rw := httptest.NewRecorder()

		h.ServeHTTP(rw, req)
		if got, want := rw.Code, 405; got != want {
			t.Errorf("%d. rw.Code: got %d, want %d", i, got, want)
		}
		if got, want := rw.Header().Get("Allow"), "GET"; got != want {
			t.Errorf("%d. rw.Header().Get(%q): got %q, want %q", i, "Allow", got, want)
		}
	}
}

func TestHandlerServeHTTP(t *testing.T) {
	v := NewVerification()
	h := NewHandler(v)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req)
	if err != nil {
		t.Fatalf("martian.TestContext(): got %v, want no error", err)
	}
	defer remove()

	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	ebs := make([]*ErrorBuilder, 3)
	for i := 0; i < len(ebs); i++ {
		kind := fmt.Sprintf("verifier.%d", i)

		eb := NewError(kind).
			Request(req).
			Actual(fmt.Sprintf("actual %d", i)).
			Expected(fmt.Sprintf("expected %d", i))

		Verify(ctx, eb)

		ebs[i] = eb
	}

	res := proxyutil.NewResponse(200, nil, req)
	if err := v.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if got, want := rw.Code, 200; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}

	ejs := &errorsJSON{}
	if err := json.Unmarshal(rw.Body.Bytes(), ejs); err != nil {
		t.Fatalf("json.Unmarshal(): got %v, want no error", err)
	}

	if got, want := len(ejs.Errors), 3; got != want {
		t.Fatalf("json.Errors: got %d errors, want %d errors", got, want)
	}

	for i, ej := range ejs.Errors {
		verr, ok := ebs[i].Error()
		if !ok {
			t.Fatalf("ebs[%d].Error(): got !ok, want ok", i)
		}

		if got, want := ej.Kind, verr.Kind; got != want {
			t.Errorf("json.Kind: got %q, want %q", got, want)
		}
		if got, want := ej.URL, verr.URL; got != want {
			t.Errorf("json.URL: got %q, want %q", got, want)
		}
		if got, want := ej.Actual, verr.Actual; got != want {
			t.Errorf("json.Actual: got %q, want %q", got, want)
		}
		if got, want := ej.Expected, verr.Expected; got != want {
			t.Errorf("json.Expected: got %q, want %q", got, want)
		}
	}
}

func TestResetHandlerServeHTTPUnsupportedMethod(t *testing.T) {
	h := NewResetHandler(NewVerification())

	for i, m := range []string{"GET", "PUT", "DELETE"} {
		req, err := http.NewRequest(m, "http://example.com", nil)
		if err != nil {
			t.Fatalf("%d. http.NewRequest(): got %v, want no error", i, err)
		}
		rw := httptest.NewRecorder()

		h.ServeHTTP(rw, req)
		if got, want := rw.Code, 405; got != want {
			t.Errorf("%d. rw.Code: got %d, want %d", i, got, want)
		}
		if got, want := rw.Header().Get("Allow"), "POST"; got != want {
			t.Errorf("%d. rw.Header().Get(%q): got %q, want %q", i, "Allow", got, want)
		}
	}
}

func TestResetHandlerServeHTTP(t *testing.T) {
	v := NewVerification()
	h := NewResetHandler(v)

	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req)
	if err != nil {
		t.Fatalf("martian.TestContext(): got %v, want no error", err)
	}
	defer remove()

	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	var reset bool
	eb := NewError("reset.Verifier").
		Request(req).
		Resets(func() { reset = true })

	Verify(ctx, eb)

	res := proxyutil.NewResponse(200, nil, req)
	if err := v.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if got, want := rw.Code, 204; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}

	if !reset {
		t.Error("reset: got false, want true")
	}
}
