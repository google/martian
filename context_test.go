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
	"testing"
)

func TestContexts(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx := TestContext(req, nil, nil)
	ctx.Set("key", "value")
	got, ok := ctx.Get("key")
	if !ok {
		t.Errorf("ctx.Get(%q): got !ok, want ok", "key")
	}
	if want := "value"; got != want {
		t.Errorf("ctx.Get(%q): got %q, want %q", "key", got, want)
	}

	ctx.SkipRoundTrip()
	if !ctx.SkippingRoundTrip() {
		t.Error("ctx.SkippingRoundTrip(): got false, want true")
	}

	ctx.SkipLogging()
	if !ctx.SkippingLogging() {
		t.Error("ctx.SkippingLogging(): got false, want true")
	}

	s := ctx.Session()
	s.MarkSecure()
	if !s.IsSecure() {
		t.Error("s.IsSecure(): got false, want true")
	}

	s.Set("key", "value")
	got, ok = s.Get("key")
	if !ok {
		t.Errorf("s.Get(%q): got !ok, want ok", "key")
	}
	if want := "value"; got != want {
		t.Errorf("s.Get(%q): got %q, want %q", "key", got, want)
	}

	ctx2 := TestContext(req, nil, nil)
	if ctx != ctx2 {
		t.Error("TestContext(): got new context, want existing context")
	}
}

func TestContextHijack(t *testing.T) {
	rc, wc := net.Pipe()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx := TestContext(req, rc, bufio.NewReadWriter(bufio.NewReader(rc), bufio.NewWriter(rc)))

	session := ctx.Session()
	if session.Hijacked() {
		t.Fatal("session.Hijacked(): got true, want false")
	}

	conn, brw, err := session.Hijack()
	if err != nil {
		t.Fatalf("session.Hijack(): got %v, want no error", err)
	}

	if !session.Hijacked() {
		t.Fatal("session.Hijacked(): got false, want true")
	}

	if _, _, err := session.Hijack(); err == nil {
		t.Fatal("session.Hijack(): got nil, want rehijack error")
	}

	go func() {
		brw.Write([]byte("test message"))
		brw.Flush()
		conn.Close()
	}()

	got, err := ioutil.ReadAll(wc)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}

	if want := []byte("test message"); !bytes.Equal(got, want) {
		t.Errorf("connection: got %q, want %q", got, want)
	}
}
