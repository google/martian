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

package martianhttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"

	_ "github.com/google/martian/header"
)

func TestIntegration(t *testing.T) {
	proxy := martian.NewProxy(nil)
	proxy.RoundTripper = martian.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return proxyutil.NewResponse(200, nil, req), nil
	})

	m := NewModifier()

	proxy.SetRequestModifier(m)
	proxy.SetResponseModifier(m)

	mux := http.NewServeMux()
	mux.Handle("/martian/modifiers", m)
	mux.Handle("/", proxy)

	s := httptest.NewServer(mux)
	defer s.Close()

	msg := []byte(`
	{
		"header.Modifier": {
      "scope": ["request", "response"],
			"name": "Martian-Test",
			"value": "true"
		}
	}`)

	req, err := http.NewRequest("POST", s.URL+"/martian/modifiers", bytes.NewBuffer(msg))
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Content-Type", "application/json")

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatalf("url.Parse(%s): got %v, want no error", s.URL, err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(u),
	}

	res, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("transport.RoundTrip(%s): got %v, want no error", req.URL, err)
	}
	res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}

	url := "http://example.com"
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("http.NewRequest(..., %q, nil): got %v, want no error", url, err)
	}
	req.Header.Set("Connection", "close")

	res, err = transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("transport.RoundTrip(%q): got %v, want no error", url, err)
	}
	res.Body.Close()

	if got, want := res.Header.Get("Martian-Test"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Martian-Test", got, want)
	}
}
