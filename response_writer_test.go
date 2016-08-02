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

package martian

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	sconn, cconn := net.Pipe()

	go func() {
		brw := bufio.NewReadWriter(bufio.NewReader(sconn), bufio.NewWriter(sconn))
		defer brw.Flush()

		rw := newResponseWriter(sconn, brw, true)
		defer rw.Close()

		rw.Header().Set("Martian-Response", "true")
		rw.Header().Set("Content-Length", "12")
		rw.Write([]byte("test "))

		// This will be ignored.
		rw.WriteHeader(400)

		rw.Write([]byte("content"))
	}()

	res, err := http.ReadResponse(bufio.NewReader(cconn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	if !res.Close {
		t.Error("res.Close: got false, want true")
	}
	if got, want := res.ContentLength, int64(12); got != want {
		t.Errorf("res.ContentLength: got %d, want %d", got, want)
	}
	if got, want := res.Header.Get("Martian-Response"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Martian-Response", got, want)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}

	if want := []byte("test content"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}

func TestResponseWriterChunkedEncoding(t *testing.T) {
	sconn, cconn := net.Pipe()

	go func() {
		brw := bufio.NewReadWriter(bufio.NewReader(sconn), bufio.NewWriter(sconn))
		defer brw.Flush()

		rw := newResponseWriter(sconn, brw, true)
		defer rw.Close()

		rw.Header().Set("Martian-Response", "true")
		rw.Write([]byte("test content"))
	}()

	res, err := http.ReadResponse(bufio.NewReader(cconn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	if !res.Close {
		t.Error("res.Close: got false, want true")
	}
	if got, want := res.TransferEncoding, []string{"chunked"}; !reflect.DeepEqual(got, want) {
		t.Errorf("res.TransferEncoding: got %v, want %v", got, want)
	}
	if got, want := res.Header.Get("Martian-Response"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Martian-Response", got, want)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}

	if want := []byte("test content"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}
