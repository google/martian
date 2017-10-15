// Copyright 2017 Google Inc. All rights reserved.
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

package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
)

const testBucket = "test_bucket"

func newTempFile(t *testing.T) *os.File {
	t.Helper()

	f, err := ioutil.TempFile("", fmt.Sprintf("%s_cache.db", t.Name()))
	if err != nil {
		t.Fatalf("ioutil.TempFile(): got error %v, want no error", err)
	}
	return f
}

func newModifier(t *testing.T, filepath string, update, replay, hermetic bool) *modifier {
	t.Helper()

	mod, err := NewModifier(filepath, testBucket, update, replay, hermetic)
	if err != nil {
		t.Fatalf("NewModifier(%q, %q, %v, %v, %v): got error %v, want no error", filepath, testBucket, update, replay, hermetic, err)
	}
	return mod.(*modifier)
}

func newRequestWithContext(t *testing.T, method, url string) (*http.Request, *martian.Context, func()) {
	t.Helper()

	req := httptest.NewRequest(method, url, nil)
	ctx, remove, err := martian.TestContext(req, nil, nil)
	if err != nil {
		t.Fatalf("martian.TestContext(): got error %v, want no error", err)
	}
	return req, ctx, remove
}

func assertResponse(t *testing.T, res *http.Response, code int, body string) {
	t.Helper()

	if got, want := res.StatusCode, code; got != want {
		t.Errorf("res.StatusCode: got %v, want %v", got, want)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("ioutil.ReadAll(res.Body): got error %v, want no error", err)
	}
	if got, want := string(b), body; got != want {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
	if err := res.Body.Close(); err != nil {
		t.Errorf("res.Body.Close(): got error %v, want no error", err)
	}

}

func TestCreateCacheModifier(t *testing.T) {
	f := newTempFile(t)
	defer os.RemoveAll(f.Name())

	mod, err := NewModifier(f.Name(), "foo", true, true, false)
	if err != nil {
		t.Fatalf("NewModifier(): got error %v, want no error", err)
	}
	if mod == nil {
		t.Fatal("NewModifier(): mod is nil")
	}
}

func TestModifierFromJSON(t *testing.T) {
	f := newTempFile(t)
	defer os.RemoveAll(f.Name())

	fname, err := json.Marshal(f.Name())
	if err != nil {
		t.Fatalf("json.Marshal(%q): got error %v, want no error", f.Name(), err)
	}
	msg := []byte(fmt.Sprintf(`{
		"cache.Modifier": {
			"scope": ["request", "response"],
			"file": %s,
			"bucket": "foo_bucket",
			"update": true,
			"replay": true,
			"hermetic": true
		}
	}`, fname))

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got error %v, want no error", err)
	}
	if r == nil {
		t.Fatal("parse.FromJSON(): result is nil")
	}
	if r.RequestModifier() == nil {
		t.Fatal("parse.FromJSON(): result.RequestModifier() is nil")
	}
	if r.ResponseModifier() == nil {
		t.Fatal("parse.FromJSON(): result.ResponseModifier() is nil")
	}

	reqmod := r.RequestModifier().(*modifier)
	resmod := r.ResponseModifier().(*modifier)

	if reqmod.db == nil {
		t.Errorf("reqmod.db: got nil, want not nil")
	}
	if resmod.db == nil {
		t.Errorf("resmod.db: got nil, want not nil")
	}

	if got, want := reqmod.bucket, "foo_bucket"; got != want {
		t.Errorf("reqmod.bucket: got %s, want %s", got, want)
	}
	if got, want := resmod.bucket, "foo_bucket"; got != want {
		t.Errorf("resmod.bucket: got %s, want %s", got, want)
	}

	if !reqmod.update {
		t.Error("reqmod.update: got false, want true")
	}
	if !resmod.update {
		t.Error("resmod.update: got false, want true")
	}

	if !reqmod.replay {
		t.Error("reqmod.replay: got false, want true")
	}
	if !resmod.replay {
		t.Error("resmod.replay: got false, want true")
	}

	if !reqmod.hermetic {
		t.Error("reqmod.hermetic: got false, want true")
	}
	if !resmod.hermetic {
		t.Error("resmod.hermetic: got false, want true")
	}
}

func TestEmptyBucketWithNoUpdateAndNoReplayIsOK(t *testing.T) {
	f := newTempFile(t)
	defer os.RemoveAll(f.Name())

	mod, err := NewModifier(f.Name(), "", false, false, false)
	if err != nil {
		t.Fatalf("NewModifier(): got error %v, want no error", err)
	}

	req, _, remove := newRequestWithContext(t, "GET", "/hello?abc=123")
	defer remove()

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("mod.ModifyRequest(): got error %v, want no error", err)
	}

	res := proxyutil.NewResponse(http.StatusTeapot, bytes.NewReader([]byte("some tea")), req)
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("mod.ModifyResponse(): got error %v, want no error", err)
	}
}

func TestEmptyBucketWithUpdateIsNotOK(t *testing.T) {
	_, err := NewModifier("/dev/null", "", true, false, false)
	if err == nil {
		t.Fatal("NewModifier(): got no error, want error")
	}
	if msg := "bucket name cannot be empty"; !strings.Contains(err.Error(), msg) {
		t.Errorf("NewModifier(): got error %q, want error to contain %q", err, msg)
	}
}

func TestEmptyBucketWithReplayIsNotOK(t *testing.T) {
	_, err := NewModifier("/dev/null", "", false, true, false)
	if err == nil {
		t.Fatal("NewModifier(): got no error, want error")
	}
	if msg := "bucket name cannot be empty"; !strings.Contains(err.Error(), msg) {
		t.Errorf("NewModifier(): got error %q, want error to contain %q", err, msg)
	}
}

func TestNoUpdateAndNoReplay(t *testing.T) {}

func TestUpdateAndNoReplay(t *testing.T) {}

func TestNoUpdateAndReplay(t *testing.T) {}

func TestNoReplayAndHermeticGetsError(t *testing.T) {
	_, err := NewModifier("/dev/null", "rice_bucket", true, false, true)
	if err == nil {
		t.Fatal("NewModifier(): got no error, want error")
	}
	if msg := "cannot use hermetic mode if not replaying"; !strings.Contains(err.Error(), msg) {
		t.Errorf("NewModifier(): got error %q, want error to contain %q", err, msg)
	}
}

func TestHermeticCacheMiss(t *testing.T) {
	f := newTempFile(t)
	defer os.RemoveAll(f.Name())

	mod := newModifier(t, f.Name(), false, true, true)

	req, _, remove := newRequestWithContext(t, "GET", "/hello?abc=123")
	defer remove()

	err := mod.ModifyRequest(req)
	if err == nil {
		t.Fatal("mod.ModifyRequest(): got no error, want error")
	}
	if msg := "hermetic"; !strings.Contains(err.Error(), msg) {
		t.Errorf("mod.ModifyRequest(): got error %q, want error to contain %q", err, msg)
	}
}

func TestCacheAndReplay(t *testing.T) {
	f := newTempFile(t)
	defer os.RemoveAll(f.Name())

	mod := newModifier(t, f.Name(), true, true, false)

	// First roundtrip should cache response.
	req, ctx, remove := newRequestWithContext(t, "GET", "/hello?abc=123")
	defer remove()

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("mod.ModifyRequest(): got error %v, want no error", err)
	}
	if ctx.SkippingRoundTrip() {
		t.Error("mod.ModifyRequest(): got skipping round trip, want no skip round trip")
	}

	res := proxyutil.NewResponse(http.StatusTeapot, bytes.NewReader([]byte("some tea")), req)
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("mod.ModifyResponse(): got error %v, want no error", err)
	}
	assertResponse(t, res, http.StatusTeapot, "some tea")

	// Second roundtrip should replay from cache.
	req, ctx, remove = newRequestWithContext(t, "GET", "/hello?abc=123")
	defer remove()

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("mod.ModifyRequest(): got error %v, want no error", err)
	}
	if !ctx.SkippingRoundTrip() {
		t.Error("mod.ModifyRequest(): got not skipping round trip, want to skip round trip")
	}

	// Create initial dummy response.
	res = proxyutil.NewResponse(http.StatusOK, nil, req)
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("mod.ModifyResponse(): got error %v, want no error", err)
	}
	assertResponse(t, res, http.StatusTeapot, "some tea")

	// Third roundtrip should also replay from cache.
	req, ctx, remove = newRequestWithContext(t, "GET", "/hello?abc=123")
	defer remove()

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("mod.ModifyRequest(): got error %v, want no error", err)
	}
	if !ctx.SkippingRoundTrip() {
		t.Error("mod.ModifyRequest(): got not skipping round trip, want to skip round trip")
	}

	// Create initial dummy response.
	res = proxyutil.NewResponse(http.StatusOK, nil, req)
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("mod.ModifyResponse(): got error %v, want no error", err)
	}
	assertResponse(t, res, http.StatusTeapot, "some tea")

	// Fourth roundtrip should not replay from cache.
	req, ctx, remove = newRequestWithContext(t, "GET", "/hello?xyz=789")
	defer remove()

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("mod.ModifyRequest(): got error %v, want no error", err)
	}
	if ctx.SkippingRoundTrip() {
		t.Error("mod.ModifyRequest(): got skipping round trip, want no skip round trip")
	}

	res = proxyutil.NewResponse(http.StatusAccepted, bytes.NewReader([]byte("some coffee")), req)
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("mod.ModifyResponse(): got error %v, want no error", err)
	}
	assertResponse(t, res, http.StatusAccepted, "some coffee")

	// Check db file was written.
	info, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("os.Stat(%q): got error %v, want no error", f.Name(), err)
	}
	if info.Size() == 0 {
		t.Error("db file size: got empty, want not empty")
	}
}

func TestCustomCacheKey(t *testing.T) {
	f := newTempFile(t)
	defer os.RemoveAll(f.Name())

	mod := newModifier(t, f.Name(), true, true, false)

	// Custom keygen modifier that uses only the URL path as cache key.
	keyGenMod := func(req *http.Request) {
		ctx := martian.NewContext(req)
		ctx.Set(CustomKey, []byte(req.URL.Path))
	}

	// First roundtrip should cache response using custom key.
	req, _, remove := newRequestWithContext(t, "GET", "/hello?abc=123")
	defer remove()

	// Apply custom keygen.
	keyGenMod(req)

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("mod.ModifyRequest(): got error %v, want no error", err)
	}

	res := proxyutil.NewResponse(http.StatusTeapot, bytes.NewReader([]byte("some tea")), req)
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("mod.ModifyResponse(): got error %v, want no error", err)
	}
	assertResponse(t, res, http.StatusTeapot, "some tea")

	// Second roundtrip should replay from cache using custom key.
	req, _, remove = newRequestWithContext(t, "POST", "/hello?xyz=789")
	defer remove()

	// Apply custom keygen.
	keyGenMod(req)

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("mod.ModifyRequest(): got error %v, want no error", err)
	}

	// Create initial dummy response.
	res = proxyutil.NewResponse(http.StatusOK, nil, req)
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("mod.ModifyResponse(): got error %v, want no error", err)
	}
	// Should get cached response.
	assertResponse(t, res, http.StatusTeapot, "some tea")
}
