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

package header

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/martian/martiantest"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/verify"
)

func TestModifyRequest(t *testing.T) {
	f := NewFilter("mARTian-teSTInG", "true")
	f.SetRequestModifier(nil)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := f.ModifyRequest(req); err != nil {
		t.Errorf("ModifyRequest(): got %v, want no error", err)
	}

	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   false,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   true,
		},
	}

	for i, tc := range tt {
		f := NewFilter("mARTian-teSTInG", "true")
		tm := martiantest.NewModifier()
		f.SetRequestModifier(tm)

		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("%d. http.NewRequest(): got %v, want no error", i, err)
		}
		req.Header[tc.name] = tc.values

		if err := f.ModifyRequest(req); err != nil {
			t.Fatalf("%d. ModifyRequest(): got %v, want no error", i, err)
		}

		if tm.RequestModified() != tc.want {
			t.Errorf("%d. tm.RequestModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}

func TestModifyResponse(t *testing.T) {
	f := NewFilter("mARTian-teSTInG", "true")
	f.SetResponseModifier(nil)

	res := proxyutil.NewResponse(200, nil, nil)

	if err := f.ModifyResponse(res); err != nil {
		t.Errorf("ModifyResponse(): got %v, want no error", err)
	}

	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   false,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   true,
		},
	}

	for i, tc := range tt {
		f := NewFilter("mARTian-teSTInG", "true")
		tm := martiantest.NewModifier()
		f.SetResponseModifier(tm)

		res := proxyutil.NewResponse(200, nil, nil)
		res.Header[tc.name] = tc.values

		if err := f.ModifyResponse(res); err != nil {
			t.Fatalf("%d. ModifyResponse(): got %v, want no error", i, err)
		}

		if tm.ResponseModified() != tc.want {
			t.Errorf("%d. tm.ResponseModified(): got %t, want %t", i, tm.ResponseModified(), tc.want)
		}
	}
}

func TestFilterFromJSON(t *testing.T) {
	msg := []byte(`{
    "header.Filter": {
      "scope": ["request", "response"],
      "name": "Martian-Passthru",
      "value": "true",
      "modifier": {
        "header.Modifier" : {
          "scope": ["request", "response"],
          "name": "Martian-Testing",
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

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Martian-Passthru", "true")
	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := req.Header.Get("Martian-Testing"), "true"; got != want {
		t.Fatalf("req.Header.Get(%q): got %q, want %q", "Martian-Testing", got, want)
	}

	resmod := r.ResponseModifier()
	if resmod == nil {
		t.Fatal("resmod: got nil, want not nil")
	}

	res := proxyutil.NewResponse(200, nil, nil)
	res.Header.Set("Martian-Passthru", "true")
	if err := resmod.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.Header.Get("Martian-Testing"), "true"; got != want {
		t.Fatalf("res.Header.Get(%q): got %q, want %q", "Martian-Testing", got, want)
	}
}

func TestPassThroughVerifyhRequests(t *testing.T) {
	f := NewFilter("Martian-Testing", "true")
	if err := f.VerifyRequests(); err != nil {
		t.Fatalf("VerifyRequest(): got %v, want no error", err)
	}

	tv := &verify.TestVerifier{
		RequestError: errors.New("verify request failure"),
	}

	f.SetRequestModifier(tv)

	if got, want := f.VerifyRequests().Error(), "verify request failure"; got != want {
		t.Fatalf("VerifyRequests(): got %s, want %s", got, want)
	}
}

func TestPassThroughVerifyResponses(t *testing.T) {
	f := NewFilter("Martian-Testing", "true")
	if err := f.VerifyResponses(); err != nil {
		t.Fatalf("VerifyResponses(): got %v, want no error", err)
	}

	tv := &verify.TestVerifier{
		ResponseError: errors.New("verify response failure"),
	}

	f.SetResponseModifier(tv)

	if got, want := f.VerifyResponses().Error(), "verify response failure"; got != want {
		t.Fatalf("VerifyResponses(): got %s, want %s", got, want)
	}
}

func TestResets(t *testing.T) {
	f := NewFilter("Martian-Testing", "true")

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
