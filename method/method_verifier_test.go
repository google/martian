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

package method

import (
	"net/http"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

func TestVerifier(t *testing.T) {
	if _, err := NewVerifier(""); err == nil {
		t.Error("NewVerifier(): got nil, want error for empty method")
	}

	v, err := NewVerifier("GET")
	if err != nil {
		t.Fatalf("NewVerifier(): got %v, want no error", err)
	}

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req)
	if err != nil {
		t.Fatalf("martian.TestContext(): got %v, want no error", err)
	}
	defer remove()

	verify.NewContext(ctx)

	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	errs := verify.FromContext(ctx)
	if len(errs) != 0 {
		t.Errorf("verify.FromContext(): got %d errors, want 0", len(errs))
	}

	req.Method = "POST"

	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	errs = verify.FromContext(ctx)
	if len(errs) != 1 {
		t.Errorf("verify.FromContext(): got %d errors, want 1", len(errs))
	}

	ev := errs[0].Get()
	if got, want := ev.Kind, "method.Verifier"; got != want {
		t.Errorf("ev.Kind: got %q, want %q", got, want)
	}
	if got, want := ev.URL, "http://www.example.com"; got != want {
		t.Errorf("ev.URL: got %q, want %q", got, want)
	}
	if got, want := ev.Scope, verify.Request; got != want {
		t.Errorf("ev.URL: got %s, want %s", got, want)
	}
	if got, want := ev.Actual, "POST"; got != want {
		t.Errorf("ev.Actual: got %q, want %q", got, want)
	}
	if got, want := ev.Expected, "GET"; got != want {
		t.Errorf("ev.Expected: got %q, want %q", got, want)
	}
}

func TestVerifierFromJSON(t *testing.T) {
	msg := []byte(`{
    "method.Verifier": {
      "scope": ["request"],
      "method": "POST"
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

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req)
	if err != nil {
		t.Fatalf("martian.TestContext(): got %v, want no error", err)
	}
	defer remove()

	verify.NewContext(ctx)

	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	errs := verify.FromContext(ctx)
	if len(errs) != 1 {
		t.Errorf("verify.FromContext(): got %d errors, want 0", len(errs))
	}
}
