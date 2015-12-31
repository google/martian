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
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

func TestVerifier(t *testing.T) {
	match := &url.URL{
		Scheme:   "https",
		Host:     "www.example.com",
		Path:     "/test",
		RawQuery: "testing=true",
	}
	v := NewVerifier(match)

	nomatch := &url.URL{
		Scheme:   "http",
		Host:     "martian.local",
		Path:     "/prod",
		RawQuery: "testing=false",
	}

	req, err := http.NewRequest("GET", match.String(), nil)
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

	errs := verify.FromContext(ctx)
	if len(errs) != 0 {
		t.Errorf("verify.FromContext(): got %d errors, want 0", len(errs))
	}

	req.URL = nomatch

	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	errs = verify.FromContext(ctx)
	if len(errs) != 4 {
		t.Errorf("verify.FromContext(): got %d errors, want 4", len(errs))
	}

	for i, err := range errs {
		verr, ok := err.Error()
		if !ok {
			t.Fatalf("%d. err.Error(): got !ok, want ok", i)
		}

		if got, want := verr.Kind, "url.Verifier"; got != want {
			t.Errorf("%d. verr.Kind: got %q, want %q", i, got, want)
		}
		if got, want := verr.URL, nomatch.String(); got != want {
			t.Errorf("%d. verr.URL: got %q, want %q", i, got, want)
		}
		if got, want := verr.Scope, verify.Request; got != want {
			t.Errorf("%d. verr.URL: got %s, want %s", i, got, want)
		}

		var actual, expected string
		switch i {
		case 0:
			actual = nomatch.Scheme
			expected = match.Scheme
		case 1:
			actual = nomatch.Host
			expected = match.Host
		case 2:
			actual = nomatch.Path
			expected = match.Path
		case 3:
			actual = nomatch.RawQuery
			expected = match.RawQuery
		default:
			t.Fatalf("errs: unexpected error at index %d, max 4", i)
		}
		if got, want := verr.Actual, actual; got != want {
			t.Errorf("%d. ev.Actual: got %q, want %q", i, got, want)
		}
		if got, want := verr.Expected, expected; got != want {
			t.Errorf("%d. ev.Expected: got %q, want %q", i, got, want)
		}
	}
}

func TestVerifierFromJSON(t *testing.T) {
	msg := []byte(`{
    "url.Verifier": {
      "scope": ["request"],
      "scheme": "https",
      "host": "www.martian.proxy",
      "path": "/testing",
      "query": "test=true"
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

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req)
	if err != nil {
		t.Fatalf("martian.TestContext(): got %v, want no error", err)
	}
	defer remove()

	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	errs := verify.FromContext(ctx)
	if len(errs) != 4 {
		t.Errorf("verify.FromContext(): got %d errors, want 4", len(errs))
	}
}
