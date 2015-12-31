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

package status

import (
	"net/http"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/verify"
)

func TestVerifier(t *testing.T) {
	v := NewVerifier(301)

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req)
	if err != nil {
		t.Fatalf("martian.TestContext(): got %v, want no error", err)
	}
	defer remove()

	res := proxyutil.NewResponse(301, nil, req)

	if err := v.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	errs := verify.FromContext(ctx)
	if len(errs) != 0 {
		t.Errorf("verify.FromContext(): got %d errors, want 0", len(errs))
	}

	res.StatusCode = 200

	if err := v.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	errs = verify.FromContext(ctx)
	if len(errs) != 1 {
		t.Errorf("verify.FromContext(): got %d errors, want 1", len(errs))
	}

	verr, ok := errs[0].Error()
	if !ok {
		t.Fatal("errs[0].Error(): got !ok, want ok")
	}

	if got, want := verr.Kind, "status.Verifier"; got != want {
		t.Errorf("verr.Kind: got %q, want %q", got, want)
	}
	if got, want := verr.URL, "http://www.example.com"; got != want {
		t.Errorf("verr.URL: got %q, want %q", got, want)
	}
	if got, want := verr.Scope, verify.Response; got != want {
		t.Errorf("verr.URL: got %s, want %s", got, want)
	}
	if got, want := verr.Actual, "200"; got != want {
		t.Errorf("verr.Actual: got %q, want %q", got, want)
	}
	if got, want := verr.Expected, "301"; got != want {
		t.Errorf("verr.Expected: got %q, want %q", got, want)
	}
}

func TestVerifierFromJSON(t *testing.T) {
	msg := []byte(`{
    "status.Verifier": {
      "scope": ["response"],
      "statusCode": 400
    }
  }`)

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got %v, want no error", err)
	}
	resmod := r.ResponseModifier()
	if resmod == nil {
		t.Fatal("resmod: got nil, want not nil")
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

	res := proxyutil.NewResponse(200, nil, req)

	if err := resmod.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	errs := verify.FromContext(ctx)
	if len(errs) != 1 {
		t.Errorf("verify.FromContext(): got %d errors, want 0", len(errs))
	}
}
