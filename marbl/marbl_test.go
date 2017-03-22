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

package marbl

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
)

func TestMarkAPIRequestsWithHeader(t *testing.T) {
	areq, err := http.NewRequest("POST", "http://localhost:8080/configure", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(areq, nil, nil)
	if err != nil {
		t.Fatalf("TestContext(): got %v, want no error", err)
	}
	defer remove()

	ctx.APIRequest()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	_, removereq, err := martian.TestContext(req, nil, nil)
	if err != nil {
		t.Fatalf("TestContext(): got %v, want no error", err)
	}
	defer removereq()

	var b bytes.Buffer

	s := NewStream(&b)
	s.LogRequest("00000000", areq)
	s.LogRequest("00000001", req)
	s.Close()

	headers := make(map[string]string)
	reader := NewReader(&b)

	for {
		frame, err := reader.ReadFrame()
		if frame == nil {
			break
		}
		if err != nil && err != io.EOF {
			t.Fatalf("reader.ReadFrame(): got %v, want no error or io.EOF", err)
		}

		headerFrame, ok := frame.(Header)
		if !ok {
			t.Fatalf("frame.(Header): couldn't convert frame '%v' to a headerFrame", frame)
		}
		headers[headerFrame.ID+headerFrame.Name] = headerFrame.Value
	}

	apih, ok := headers["00000000:api"]
	if !ok {
		t.Errorf("headers[00000000:api]: got no such header, want :api (headers were: %v)", headers)
	}

	if got, want := apih, "true"; got != want {
		t.Errorf("headers[%q]: got %v, want %q", "00000000:api", got, want)
	}

	_, ok = headers["00000001:api"]
	if got, want := ok, false; got != want {
		t.Error("headers[00000001:api]: got :api header, want no header for non-api requests")
	}
}

func TestSendTimestampWithLogRequest(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	_, remove, err := martian.TestContext(req, nil, nil)
	if err != nil {
		t.Fatalf("TestContext(): got %v, want no error", err)
	}
	defer remove()

	var b bytes.Buffer
	s := NewStream(&b)

	before := time.Now().UnixNano() / 1000 / 1000
	s.LogRequest("Fake_Id0", req)
	s.Close()
	after := time.Now().UnixNano() / 1000 / 1000

	headers := make(map[string]string)
	reader := NewReader(&b)

	for {
		frame, err := reader.ReadFrame()
		if frame == nil {
			break
		}
		if err != nil && err != io.EOF {
			t.Fatalf("reader.ReadFrame(): got %v, want no error or io.EOF", err)
		}

		headerFrame, ok := frame.(Header)
		if !ok {
			t.Fatalf("frame.(Header): couldn't convert frame '%v' to a headerFrame", frame)
		}
		headers[headerFrame.Name] = headerFrame.Value
	}

	timestr, ok := headers[":timestamp"]
	if !ok {
		t.Fatalf("headers[:timestamp]: got no such header, want :timestamp (headers were: %v)", headers)
	}
	ts, err := strconv.ParseInt(timestr, 10, 64)
	if err != nil {
		t.Fatalf("strconv.ParseInt: got %s, want no error. Invalidly formatted timestamp ('%s')", err, timestr)
	}
	if ts < before || ts > after {
		t.Fatalf("headers[:timestamp]: got %d, want timestamp between %d and %d", ts, before, after)
	}
}

func TestSendTimestampWithLogResponse(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	_, remove, err := martian.TestContext(req, nil, nil)
	if err != nil {
		t.Fatalf("TestContext(): got %v, want no error", err)
	}
	defer remove()

	res := proxyutil.NewResponse(200, nil, req)
	var b bytes.Buffer
	s := NewStream(&b)

	before := time.Now().UnixNano() / 1000 / 1000
	s.LogResponse("Fake_Id1", res)
	s.Close()
	after := time.Now().UnixNano() / 1000 / 1000

	headers := make(map[string]string)
	reader := NewReader(&b)

	for {
		frame, err := reader.ReadFrame()
		if frame == nil {
			break
		}
		if err != nil && err != io.EOF {
			t.Fatalf("reader.ReadFrame(): got %v, want no error or io.EOF", err)
		}

		headerFrame, ok := frame.(Header)
		if !ok {
			t.Fatalf("frame.(Header): couldn't convert frame '%v' to a headerFrame", frame)
		}
		headers[headerFrame.Name] = headerFrame.Value
	}

	timestr, ok := headers[":timestamp"]
	if !ok {
		t.Fatalf("headers[:timestamp]: got no such header, want :timestamp (headers were: %v)", headers)
	}
	ts, err := strconv.ParseInt(timestr, 10, 64)
	if err != nil {
		t.Fatalf("strconv.ParseInt: got %s, want no error. Invalidly formatted timestamp ('%s')", err, timestr)
	}
	if ts < before || ts > after {
		t.Fatalf("headers[:timestamp]: got %d, want timestamp between %d and %d (headers were: %v)", ts, before, after, headers)
	}
}
