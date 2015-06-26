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

package proxyutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNewResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	res := NewResponse(200, nil, req)
	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	if got, want := res.Status, "200 OK"; got != want {
		t.Errorf("res.Status: got %q, want %q", got, want)
	}
	if got, want := res.Proto, "HTTP/1.1"; got != want {
		t.Errorf("res.Proto: got %q, want %q", got, want)
	}
	if got, want := res.ProtoMajor, 1; got != want {
		t.Errorf("res.ProtoMajor: got %d, want %d", got, want)
	}
	if got, want := res.ProtoMinor, 1; got != want {
		t.Errorf("res.ProtoMinor: got %d, want %d", got, want)
	}
	if res.Header == nil {
		t.Error("res.Header: got nil, want header")
	}
	if _, ok := res.Body.(io.ReadCloser); !ok {
		t.Error("res.Body.(io.ReadCloser): got !ok, want ok")
	}
	if got, want := res.Request, req; got != want {
		t.Errorf("res.Request: got %v, want %v", got, want)
	}
}

func TestNewErrorResponse(t *testing.T) {
	err := fmt.Errorf("response error")
	res := NewErrorResponse(502, err, nil)

	if got, want := res.StatusCode, 502; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	if got, want := res.Status, "502 Bad Gateway"; got != want {
		t.Errorf("res.Status: got %q, want %q", got, want)
	}
	if got, want := res.Header.Get("Content-Type"), "text/plain; charset=utf-8"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Content-Type", got, want)
	}
	if got, want := res.ContentLength, int64(len("response error")); got != want {
		t.Errorf("res.ContentLength: got %d, want %d", got, want)
	}
	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	if want := []byte("response error"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}
