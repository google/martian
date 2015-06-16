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
	"reflect"
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

func TestRemoveHopByHopHeaders(t *testing.T) {
	header := http.Header{
		// Additional hop-by-hop headers are listed in the
		// Connection header.
		"Connection": []string{
			"X-Connection",
			"X-Hop-By-Hop, close",
		},

		// RFC hop-by-hop headers.
		"Keep-Alive":		[]string{},
		"Proxy-Authenticate":	[]string{},
		"Proxy-Authorization":	[]string{},
		"Te":			[]string{},
		"Trailer":		[]string{},
		"Transfer-Encoding":	[]string{},
		"Upgrade":		[]string{},

		// Hop-by-hop headers listed in the Connection header.
		"X-Connection":	[]string{},
		"X-Hop-By-Hop":	[]string{},

		// End-to-end header that should not be removed.
		"X-End-To-End":	[]string{},
	}

	RemoveHopByHopHeaders(header)

	if got, want := len(header), 1; got != want {
		t.Fatalf("len(header): got %d, want %d", got, want)
	}
	if _, ok := header["X-End-To-End"]; !ok {
		t.Errorf("header[%q]: got !ok, want ok", "X-End-To-End")
	}
}

func TestSetForwardHeaders(t *testing.T) {
	xfp := "X-Forwarded-Proto"
	xff := "X-Forwarded-For"

	req, err := http.NewRequest("GET", "http://martian.local", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.RemoteAddr = "10.0.0.1:8112"
	SetForwardedHeaders(req)

	if got, want := req.Header.Get(xfp), "http"; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", xfp, got, want)
	}
	if got, want := req.Header.Get(xff), "10.0.0.1"; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", xff, got, want)
	}

	// Test with existing X-Forwarded-For.
	req.RemoteAddr = "12.12.12.12"
	SetForwardedHeaders(req)

	if got, want := req.Header.Get(xff), "10.0.0.1, 12.12.12.12"; got != want {
		t.Errorf("req.Header.Get(%q): got %q, want %q", xff, got, want)
	}
}

func TestSetViaHeader(t *testing.T) {
	header := http.Header{}

	SetViaHeader(header, "1.1 martian")
	if got, want := header.Get("Via"), "1.1 martian"; got != want {
		t.Errorf("header.Get(%q): got %q, want %q", "Via", got, want)
	}

	header.Set("Via", "1.0 alpha")
	SetViaHeader(header, "1.1 martian")
	if got, want := header.Get("Via"), "1.0 alpha, 1.1 martian"; got != want {
		t.Errorf("header.Get(%q): got %q, want %q", "Via", got, want)
	}
}

func TestBadFramingMultipleContentLengths(t *testing.T) {
	header := http.Header{
		"Content-Length": []string{"42", "42, 42"},
	}

	if err := FixBadFraming(header); err != nil {
		t.Errorf("FixBadFraming(): got %v, want no error", err)
	}
	if got, want := header["Content-Length"], []string{"42"}; !reflect.DeepEqual(got, want) {
		t.Errorf("header[%q]: got %v, want %v", "Content-Length", got, want)
	}

	header["Content-Length"] = []string{"42", "32, 42"}
	if got, want := FixBadFraming(header), ErrBadFraming; got != want {
		t.Errorf("FixBadFraming(): got %v, want %v", got, want)
	}
}

func TestBadFramingTransferEncodingAndContentLength(t *testing.T) {
	header := http.Header{
		"Transfer-Encoding":	[]string{"gzip, chunked"},
		"Content-Length":	[]string{"42"},
	}

	if err := FixBadFraming(header); err != nil {
		t.Errorf("FixBadFraming(): got %v, want no error", err)
	}
	if _, ok := header["Content-Length"]; ok {
		t.Fatalf("header[%q]: got ok, want !ok", "Content-Length")
	}

	header.Set("Transfer-Encoding", "gzip, identity")
	header.Del("Content-Length")
	if got, want := FixBadFraming(header), ErrBadFraming; got != want {
		t.Errorf("FixBadFraming(): got %v, want %v", got, want)
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
