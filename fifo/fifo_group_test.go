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
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/verify"

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
	if err := reqmod.ModifyRequest(martian.NewContext(), req); err != nil {
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
	if err := resmod.ModifyResponse(martian.NewContext(), res); err != nil {
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
	mg := NewGroup()

	modifierRun := false
	f := func(*martian.Context, *http.Request) error {
		modifierRun = true
		return nil
	}
	mg.AddRequestModifier(martian.RequestModifierFunc(f))

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := mg.ModifyRequest(martian.NewContext(), req); err != nil {
		t.Fatalf("mg.ModifyRequest(): got %v, want no error", err)
	}
	if !modifierRun {
		t.Error("modifierRun: got false, want true")
	}
}

func TestModifyRequestHaltsOnError(t *testing.T) {
	mg := NewGroup()
	errHalt := errors.New("halt modifier chain")
	f := func(*martian.Context, *http.Request) error {
		return errHalt
	}
	mg.AddRequestModifier(martian.RequestModifierFunc(f))

	f = func(*martian.Context, *http.Request) error {
		t.Fatal("ModifyRequest(): got called, want skipped")
		return nil
	}
	mg.AddRequestModifier(martian.RequestModifierFunc(f))

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := mg.ModifyRequest(martian.NewContext(), req); err != errHalt {
		t.Fatalf("mg.ModifyRequest(): got %v, want %v", err, errHalt)
	}
}

func TestModifyResponse(t *testing.T) {
	mg := NewGroup()

	modifierRun := false
	f := func(*martian.Context, *http.Response) error {
		modifierRun = true
		return nil
	}
	mg.AddResponseModifier(martian.ResponseModifierFunc(f))

	res := proxyutil.NewResponse(200, nil, nil)
	if err := mg.ModifyResponse(martian.NewContext(), res); err != nil {
		t.Fatalf("mg.ModifyResponse(): got %v, want no error", err)
	}
	if !modifierRun {
		t.Error("modifierRun: got false, want true")
	}
}

func TestModifyResponseHaltsOnError(t *testing.T) {
	mg := NewGroup()
	errHalt := errors.New("halt modifier chain")
	f := func(*martian.Context, *http.Response) error {
		return errHalt
	}
	mg.AddResponseModifier(martian.ResponseModifierFunc(f))

	f = func(*martian.Context, *http.Response) error {
		t.Fatal("ModifyResponse(): got called, want skipped")
		return nil
	}
	mg.AddResponseModifier(martian.ResponseModifierFunc(f))

	res := proxyutil.NewResponse(200, nil, nil)
	if err := mg.ModifyResponse(martian.NewContext(), res); err != errHalt {
		t.Fatalf("mg.ModifyResponse(): got %v, want %v", err, errHalt)
	}
}

func TestVerifyRequests(t *testing.T) {
	mg := NewGroup()

	if err := mg.VerifyRequests(); err != nil {
		t.Fatalf("VerifyRequest(): got %v, want no error", err)
	}

	errs := []error{}
	for i := 0; i < 3; i++ {
		err := fmt.Errorf("%d. verify request failure", i)

		tv := &verify.TestVerifier{
			RequestError: err,
		}
		mg.AddRequestModifier(tv)

		errs = append(errs, err)
	}

	merr, ok := mg.VerifyRequests().(*verify.MultiError)
	if !ok {
		t.Fatal("VerifyRequests(): got nil, want *verify.MultiError")
	}

	if !reflect.DeepEqual(merr.Errors(), errs) {
		t.Errorf("merr.Errors(): got %v, want %v", merr.Errors(), errs)
	}
}

func TestVerifyResponses(t *testing.T) {
	mg := NewGroup()

	if err := mg.VerifyResponses(); err != nil {
		t.Fatalf("VerifyResponses(): got %v, want no error", err)
	}

	errs := []error{}
	for i := 0; i < 3; i++ {
		err := fmt.Errorf("%d. verify responses failure", i)

		tv := &verify.TestVerifier{
			ResponseError: err,
		}
		mg.AddResponseModifier(tv)

		errs = append(errs, err)
	}

	merr, ok := mg.VerifyResponses().(*verify.MultiError)
	if !ok {
		t.Fatal("VerifyResponses(): got nil, want *verify.MultiError")
	}

	if !reflect.DeepEqual(merr.Errors(), errs) {
		t.Errorf("merr.Errors(): got %v, want %v", merr.Errors(), errs)
	}
}

func TestResets(t *testing.T) {
	mg := NewGroup()

	for i := 0; i < 3; i++ {
		tv := &verify.TestVerifier{
			RequestError:  fmt.Errorf("%d. verify request error", i),
			ResponseError: fmt.Errorf("%d. verify response error", i),
		}
		mg.AddRequestModifier(tv)
		mg.AddResponseModifier(tv)
	}

	if err := mg.VerifyRequests(); err == nil {
		t.Fatal("VerifyRequests(): got nil, want error")
	}
	if err := mg.VerifyResponses(); err == nil {
		t.Fatal("VerifyResponses(): got nil, want error")
	}

	mg.ResetRequestVerifications()
	mg.ResetResponseVerifications()

	if err := mg.VerifyRequests(); err != nil {
		t.Errorf("VerifyRequests(): got %v, want no error", err)
	}
	if err := mg.VerifyResponses(); err != nil {
		t.Errorf("VerifyResponses(): got %v, want no error", err)
	}
}
