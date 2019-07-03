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

package method

import (
	"net/http"
	"testing"

	"github.com/google/martian/proxyutil"

	"github.com/google/martian/martiantest"
)

func TestFilterModifyRequest(t *testing.T) {
	tt := []struct {
		method string
		want   bool
	}{
		{
			method: "GET",
			want:   true,
		},
		{
			method: "POST",
			want:   false,
		},
		{
			method: "DELETE",
			want:   false,
		},
		{
			method: "CONNECT",
			want:   false,
		},
	}

	for i, tc := range tt {
		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("%d. NewRequest(): got %v, want no error", i, err)
		}

		mod := NewFilter(tc.method)
		tm := martiantest.NewModifier()
		mod.SetRequestModifier(tm)

		if err := mod.ModifyRequest(req); err != nil {
			t.Fatalf("%d. ModifyRequest(): got %q, want no error", i, err)
		}

		if tm.RequestModified() != tc.want {
			t.Errorf("%d. tm.RequestModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}

func TestFilterModifyResponse(t *testing.T) {
	tt := []struct {
		method string
		want   bool
	}{
		{
			method: "GET",
			want:   true,
		},
		{
			method: "POST",
			want:   false,
		},
		{
			method: "DELETE",
			want:   false,
		},
		{
			method: "CONNECT",
			want:   false,
		},
	}

	for i, tc := range tt {
		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("%d. NewRequest(): got %v, want no error", i, err)
		}
		res := proxyutil.NewResponse(200, nil, req)

		mod := NewFilter(tc.method)
		tm := martiantest.NewModifier()
		mod.SetResponseModifier(tm)

		if err := mod.ModifyResponse(res); err != nil {
			t.Fatalf("%d. ModifyResponse(): got %q, want no error", i, err)
		}

		if tm.ResponseModified() != tc.want {
			t.Errorf("%d. tm.ResponseModified(): got %t, want %t", i, tm.ResponseModified(), tc.want)
		}
	}

}
