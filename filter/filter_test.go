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

package filter

import (
	"net/http"
	"testing"

	"github.com/google/martian/header"
	"github.com/google/martian/martiantest"
	"github.com/google/martian/proxyutil"
)

func TestRequestWhenTrueCondition(t *testing.T) {
	hm := header.NewMatcher("Martian-Testing", "true")

	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   false,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   true,
		},
	}

	for i, tc := range tt {
		tm := martiantest.NewModifier()

		f := New()
		f.SetRequestCondition(hm)
		f.RequestWhenTrue(tm)

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		req.Header[tc.name] = tc.values

		if err := f.ModifyRequest(req); err != nil {
			t.Fatalf("%d. ModifyRequest(): got %v, want no error", i, err)
		}

		if tm.RequestModified() != tc.want {
			t.Errorf("%d. tm.RequestModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}

func TestRequestWhenFalse(t *testing.T) {
	hm := header.NewMatcher("Martian-Testing", "true")
	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   true,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   false,
		},
	}

	for i, tc := range tt {
		tm := martiantest.NewModifier()

		f := New()
		f.SetRequestCondition(hm)
		f.RequestWhenFalse(tm)

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		req.Header[tc.name] = tc.values

		if err := f.ModifyRequest(req); err != nil {
			t.Fatalf("%d. ModifyRequest(): got %v, want no error", i, err)
		}

		if tm.RequestModified() != tc.want {
			t.Errorf("%d. tm.RequestModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}

func TestResponseWhenTrue(t *testing.T) {
	hm := header.NewMatcher("Martian-Testing", "true")

	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   false,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   true,
		},
	}

	for i, tc := range tt {
		tm := martiantest.NewModifier()

		f := New()
		f.SetResponseCondition(hm)
		f.ResponseWhenTrue(tm)

		res := proxyutil.NewResponse(200, nil, nil)

		res.Header[tc.name] = tc.values

		if err := f.ModifyResponse(res); err != nil {
			t.Fatalf("%d. ModifyResponse(): got %v, want no error", i, err)
		}

		if tm.ResponseModified() != tc.want {
			t.Errorf("%d. tm.ResponseModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}

func TestResponseWhenFalse(t *testing.T) {
	hm := header.NewMatcher("Martian-Testing", "true")

	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   true,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   false,
		},
	}

	for i, tc := range tt {
		tm := martiantest.NewModifier()

		f := New()
		f.SetResponseCondition(hm)
		f.ResponseWhenFalse(tm)

		res := proxyutil.NewResponse(200, nil, nil)

		res.Header[tc.name] = tc.values

		if err := f.ModifyResponse(res); err != nil {
			t.Fatalf("%d. ModifyResponse(): got %v, want no error", i, err)
		}

		if tm.ResponseModified() != tc.want {
			t.Errorf("%d. tm.ResponseModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}
