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

package trafficshape

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	tt := []struct {
		message string
		query   string
		status  int

		latency time.Duration
		up      int64
		down    int64
	}{
		{
			message: "latency, up, and down are valid",
			query:   "?latency=10ms&down=1000&up=1000",
			status:  200,
			latency: 10 * time.Millisecond,
			up:      1000,
			down:    1000,
		},
		{
			message: "empty query string",
			query:   "",
			status:  200,
			latency: time.Duration(0),
			up:      defaultBitrate,
			down:    defaultBitrate,
		},
		{
			message: "invalid latency",
			query:   "?latency=ten-millis",
			status:  400,
			up:      defaultBitrate,
			down:    defaultBitrate,
		},
		{
			message: "valid latency, invalid upstream",
			query:   "?latency=10ms&up=ten",
			status:  400,
			latency: 10 * time.Millisecond,
			up:      defaultBitrate,
			down:    defaultBitrate,
		},
		{
			message: "valid latency, upstream, invalid downstream",
			query:   "?latency=10ms&up=1000&down=ten",
			status:  400,
			latency: 10 * time.Millisecond,
			up:      1000,
			down:    defaultBitrate,
		},
	}

	for i, tc := range tt {
		t.Logf("case %d: %s", i+1, tc.message)

		tsl := NewListener(l)
		defer tsl.Close()

		h := NewHandler(tsl)

		req, err := http.NewRequest("GET", fmt.Sprintf("http://martian.proxy/%s", tc.query), nil)
		if err != nil {
			t.Fatalf("%d. http.NewRequest(): got %v, want no error", i, err)
		}
		rw := httptest.NewRecorder()

		h.ServeHTTP(rw, req)

		if got, want := rw.Code, tc.status; got != want {
			t.Errorf("%d. rw.Code: got %d, want %d", i, got, want)
		}

		if got, want := tsl.Latency(), tc.latency; got != want {
			t.Errorf("%d. tsl.Latency(): got %s, want %s", i, got, want)
		}
		if got, want := tsl.ReadBitrate(), tc.down; got != want {
			t.Errorf("%d. tsl.ReadBitrate(): got %d, want %d", i, got, want)
		}
		if got, want := tsl.WriteBitrate(), tc.up; got != want {
			t.Errorf("%d. tsl.WriteBitrate(): got %d, want %d", i, got, want)
		}
	}
}
