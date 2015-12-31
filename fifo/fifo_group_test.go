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

package fifo

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/martian/martiantest"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"

	_ "github.com/google/martian/header"
)

func TestGroupFromJSON(t *testing.T) {
	msg := []byte(`{
    "fifo.Group": {
      "scope": ["request", "response"],
      "modifiers": [
        {
          "header.Modifier": {
            "scope": ["request", "response"],
            "name": "X-Testing",
            "value": "true"
          }
        },
        {
          "header.Modifier": {
            "scope": ["request", "response"],
            "name": "Y-Testing",
            "value": "true"
          }
        }
      ]
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
	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := req.Header.Get("X-Testing"), "true"; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", "X-Testing", got, want)
	}
	if got, want := req.Header.Get("Y-Testing"), "true"; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", "Y-Testing", got, want)
	}

	resmod := r.ResponseModifier()
	if resmod == nil {
		t.Fatal("resmod: got nil, want not nil")
	}
	res := proxyutil.NewResponse(200, nil, req)
	if err := resmod.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.Header.Get("X-Testing"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "X-Testing", got, want)
	}
	if got, want := res.Header.Get("Y-Testing"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Y-Testing", got, want)
	}
}

func TestModifyRequest(t *testing.T) {
	fg := NewGroup()
	tm := martiantest.NewModifier()

	fg.AddRequestModifier(tm)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := fg.ModifyRequest(req); err != nil {
		t.Fatalf("fg.ModifyRequest(): got %v, want no error", err)
	}
	if !tm.RequestModified() {
		t.Error("tm.RequestModified(): got false, want true")
	}
}

func TestModifyRequestHaltsOnError(t *testing.T) {
	fg := NewGroup()

	reqerr := errors.New("request error")
	tm := martiantest.NewModifier()
	tm.RequestError(reqerr)
	fg.AddRequestModifier(tm)

	tm2 := martiantest.NewModifier()
	fg.AddRequestModifier(tm2)

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := fg.ModifyRequest(req); err != reqerr {
		t.Fatalf("fg.ModifyRequest(): got %v, want %v", err, reqerr)
	}

	if tm2.RequestModified() {
		t.Error("tm2.RequestModified(): got true, want false")
	}
}

func TestModifyResponse(t *testing.T) {
	fg := NewGroup()
	tm := martiantest.NewModifier()

	fg.AddResponseModifier(tm)

	res := proxyutil.NewResponse(200, nil, nil)
	if err := fg.ModifyResponse(res); err != nil {
		t.Fatalf("fg.ModifyResponse(): got %v, want no error", err)
	}
	if !tm.ResponseModified() {
		t.Error("tm.ResponseModified(): got false, want true")
	}
}

func TestModifyResponseHaltsOnError(t *testing.T) {
	fg := NewGroup()

	reserr := errors.New("request error")
	tm := martiantest.NewModifier()
	tm.ResponseError(reserr)
	fg.AddResponseModifier(tm)

	tm2 := martiantest.NewModifier()
	fg.AddResponseModifier(tm2)

	res := proxyutil.NewResponse(200, nil, nil)
	if err := fg.ModifyResponse(res); err != reserr {
		t.Fatalf("fg.ModifyResponse(): got %v, want %v", err, reserr)
	}

	if tm2.ResponseModified() {
		t.Error("tm2.ResponseModified(): got true, want false")
	}
}
