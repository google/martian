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
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/google/martian/log"
	"github.com/google/martian/proxyutil"
)

func deserializeHeaders(bs []byte) map[string]string {
	i := 8
	log.Errorf("deserializing: %v", bs)
	headers := make(map[string]string)
	var kl, vl int
	var key, val string

	for i < len(bs) {
		i += 2
		kl = int(bs[i]<<24 + bs[i+1]<<16 + bs[i+2]<<8 + bs[i+3])
		vl = int(bs[i+4]<<24 + bs[i+5]<<16 + bs[i+6]<<8 + bs[i+7])
		i += 8
		key = string(bs[i : i+kl])
		val = string(bs[i+kl : i+kl+vl])
		log.Errorf("decoded %v into %s:%s", bs[i:i+kl+vl], key, val)

		headers[key] = val
		i += kl + vl + 8
	}

	return headers
}

func TestSendTimestampWithLogRequest(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	var b bytes.Buffer
	s := NewStream(&b)

	before := time.Now().UnixNano() / 1000 / 1000
	s.LogRequest("fake_req_id", req)
	s.Close()
	after := time.Now().UnixNano() / 1000 / 1000

	ob := new(bytes.Buffer)
	ob.ReadFrom(&b)
	bs := ob.Bytes()

	headers := deserializeHeaders(bs)

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
	res := proxyutil.NewResponse(200, nil, req)
	var b bytes.Buffer
	s := NewStream(&b)

	before := time.Now().UnixNano() / 1000 / 1000
	s.LogResponse("fake_res_id", res)
	s.Close()
	after := time.Now().UnixNano() / 1000 / 1000

	ob := new(bytes.Buffer)
	ob.ReadFrom(&b)
	bs := ob.Bytes()

	headers := deserializeHeaders(bs)

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