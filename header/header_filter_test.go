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

package header

import (
	"net/http"
	"testing"

	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
)

func TestFilterFromJSON(t *testing.T) {
	msg := []byte(`{
    "header.Filter": {
      "scope": ["request", "response"],
      "name": "Martian-Passthrough",
      "value": "true",
      "modifier": {
        "header.Modifier" : {
          "scope": ["request", "response"],
          "name": "Martian-Testing",
          "value": "true"
        }
      }
    }
  }`)

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got %v, want no error", err)
	}
	reqmod := r.RequestModifier()
	if reqmod == nil {
		t.Fatal("reqmod: got nil, want not nil")
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Martian-Passthrough", "true")
	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := req.Header.Get("Martian-Testing"), "true"; got != want {
		t.Fatalf("req.Header.Get(%q): got %q, want %q", "Martian-Testing", got, want)
	}

	resmod := r.ResponseModifier()
	if resmod == nil {
		t.Fatal("resmod: got nil, want not nil")
	}

	res := proxyutil.NewResponse(200, nil, nil)
	res.Header.Set("Martian-Passthrough", "true")
	if err := resmod.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.Header.Get("Martian-Testing"), "true"; got != want {
		t.Fatalf("res.Header.Get(%q): got %q, want %q", "Martian-Testing", got, want)
	}
}
