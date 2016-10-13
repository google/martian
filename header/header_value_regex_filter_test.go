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

func TestValueRegexFilterFromJSON(t *testing.T) {
	msg := []byte(`{
    "header.RegexFilter": {
      "scope": ["request", "response"],
	    "header": "X-Forwarded-Url",
	    "regex": ".*/test",
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

	resmod := r.ResponseModifier()
	if resmod == noop {
		t.Fatalf("r.ResponseModifier: got nil, want not nil")
	}

	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	req.Header.Add("X-Forwarded-Url", "http://yahoo.com/test")
	res := proxyutil.NewResponse(200, nil, req)

	resmod.ModifyResponse(res)

	if got, want := res.Header.Get("Martian-Testing"), "true"; got != want {
		t.Fatalf("res.Header.Get(%q): got %v, want %v", "Martian-Testing", got, want)
	}
}
