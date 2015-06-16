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

package body

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
)

func TestBodyModifier(t *testing.T) {
	mod, err := NewModifier([]byte("text"), "text/plain")
	if err != nil {
		t.Fatalf("NewModifier(): got %v, want no error", err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}

	ctx := martian.NewContext()
	if err := mod.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if !ctx.SkipRoundTrip {
		t.Error("ctx.SkipRoundTrip: got false, want true")
	}

	res := proxyutil.NewResponse(200, nil, nil)
	res.Header.Set("Content-Encoding", "gzip")

	if err := mod.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	if got, want := res.Header.Get("Content-Type"), "text/plain"; got != want {
		t.Errorf("res.Header.Get(%q): got %v, want %v", "Content-Type", got, want)
	}
	if got, want := res.ContentLength, int64(len([]byte("text"))); got != want {
		t.Errorf("res.ContentLength: got %d, want %d", got, want)
	}
	if got, want := res.Header.Get("Content-Encoding"), ""; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Content-Encoding", got, want)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	res.Body.Close()

	if want := []byte("text"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}

func TestModifierFromJSON(t *testing.T) {
	rawMsg := `
	{
	  "body.Modifier":{
		  "scope": ["response"],
  	  "contentType": "text/plain",
	  	"body": %q
    }
	}
	`

	payload := base64.StdEncoding.EncodeToString([]byte("data"))
	msg := []byte(fmt.Sprintf(rawMsg, payload))

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got %v, want no error", err)
	}

	resmod := r.ResponseModifier()

	if resmod == nil {
		t.Fatalf("resmod: got nil, want not nil")
	}

	res := proxyutil.NewResponse(200, nil, nil)
	if err := resmod.ModifyResponse(martian.NewContext(), res); err != nil {
		t.Fatalf("resmod.ModifyResponse(): got %v, want no error", err)
	}

	if got, want := res.Header.Get("Content-Type"), "text/plain"; got != want {
		t.Errorf("res.Header.Get(%q): got %v, want %v", "Content-Type", got, want)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	res.Body.Close()

	if want := []byte("data"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}
