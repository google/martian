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

package priority

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"

	// Import to register header.Modifier with JSON parser.
	_ "github.com/google/martian/header"
)

func TestPriorityGroupModifyRequest(t *testing.T) {
	var priorities []int64

	pg := NewGroup()
	f := func(*martian.Context, *http.Request) error {
		priorities = append(priorities, 50)
		return nil
	}
	pg.AddRequestModifier(martian.RequestModifierFunc(f), 50)

	f = func(*martian.Context, *http.Request) error {
		priorities = append(priorities, 100)
		return nil
	}
	pg.AddRequestModifier(martian.RequestModifierFunc(f), 100)

	f = func(*martian.Context, *http.Request) error {
		priorities = append(priorities, 75)
		return nil
	}

	// Functions are not directly comparable, so we must wrap in a
	// type that is.
	m := &struct{ martian.RequestModifier }{martian.RequestModifierFunc(f)}
	if err := pg.RemoveRequestModifier(m); err != ErrModifierNotFound {
		t.Fatalf("RemoveRequestModifier(): got %v, want ErrModifierNotFound", err)
	}
	pg.AddRequestModifier(m, 75)
	if err := pg.RemoveRequestModifier(m); err != nil {
		t.Fatalf("RemoveRequestModifier(): got %v, want no error", err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := pg.ModifyRequest(martian.NewContext(), req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := priorities, []int64{100, 50}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reflect.DeepEqual(%v, %v): got false, want true", got, want)
	}
}

func TestPriorityGroupModifyRequestHaltsOnError(t *testing.T) {
	pg := NewGroup()

	errHalt := errors.New("modifier chain halted")
	f := func(*martian.Context, *http.Request) error {
		return errHalt
	}
	pg.AddRequestModifier(martian.RequestModifierFunc(f), 100)

	f = func(*martian.Context, *http.Request) error {
		t.Fatal("ModifyRequest(): got called, want skipped")
		return nil
	}
	pg.AddRequestModifier(martian.RequestModifierFunc(f), 75)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := pg.ModifyRequest(martian.NewContext(), req); err != errHalt {
		t.Fatalf("ModifyRequest(): got %v, want errHalt", err)
	}
}

func TestPriorityGroupModifyResponse(t *testing.T) {
	var priorities []int64

	pg := NewGroup()
	f := func(*martian.Context, *http.Response) error {
		priorities = append(priorities, 50)
		return nil
	}
	pg.AddResponseModifier(martian.ResponseModifierFunc(f), 50)

	f = func(*martian.Context, *http.Response) error {
		priorities = append(priorities, 100)
		return nil
	}
	pg.AddResponseModifier(martian.ResponseModifierFunc(f), 100)

	f = func(*martian.Context, *http.Response) error {
		priorities = append(priorities, 75)
		return nil
	}

	// Functions are not directly comparable, so we must wrap in a
	// type that is.
	m := &struct{ martian.ResponseModifier }{martian.ResponseModifierFunc(f)}
	if err := pg.RemoveResponseModifier(m); err != ErrModifierNotFound {
		t.Fatalf("RemoveResponseModifier(): got %v, want ErrModifierNotFound", err)
	}
	pg.AddResponseModifier(m, 75)
	if err := pg.RemoveResponseModifier(m); err != nil {
		t.Fatalf("RemoveResponseModifier(): got %v, want no error", err)
	}

	res := proxyutil.NewResponse(200, nil, nil)
	if err := pg.ModifyResponse(martian.NewContext(), res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if got, want := priorities, []int64{100, 50}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reflect.DeepEqual(%v, %v): got false, want true", got, want)
	}
}

func TestPriorityGroupModifyResponseHaltsOnError(t *testing.T) {
	pg := NewGroup()

	errHalt := errors.New("modifier chain halted")
	f := func(*martian.Context, *http.Response) error {
		return errHalt
	}
	pg.AddResponseModifier(martian.ResponseModifierFunc(f), 100)

	f = func(*martian.Context, *http.Response) error {
		t.Fatal("ModifyResponse(): got called, want skipped")
		return nil
	}
	pg.AddResponseModifier(martian.ResponseModifierFunc(f), 75)

	res := proxyutil.NewResponse(200, nil, nil)
	if err := pg.ModifyResponse(martian.NewContext(), res); err != errHalt {
		t.Fatalf("ModifyResponse(): got %v, want errHalt", err)
	}
}

func TestGroupFromJSON(t *testing.T) {
	msg := []byte(`{
    "priority.Group": {
    "scope": ["request", "response"],
    "modifiers": [
      {
        "priority": 100,
        "modifier": {
          "header.Modifier": {
            "scope": ["request", "response"],
            "name": "X-Testing",
            "value": "true"
          }
        }
      },
      {
        "priority": 0,
        "modifier": {
          "header.Modifier": {
            "scope": ["request", "response"],
            "name": "Y-Testing",
            "value": "true"
          }
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
