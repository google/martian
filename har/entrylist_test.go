// Copyright 2020 Google Inc. All rights reserved.
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

package har

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/martian/v3"
)

func TestEntryList(t *testing.T) {
	ids := make([]string, 3)
	urls := make([]string, 3)

	logger := NewLogger()

	urls[0] = "http://0.example.com/path"
	urls[1] = "http://1.example.com/path"
	urls[2] = "http://2.example.com/path"

	for idx, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		_, remove, err := martian.TestContext(req, nil, nil)
		if err != nil {
			t.Fatalf("martian.TestContext(): got %v, want no error", err)
		}
		defer remove()

		if err := logger.ModifyRequest(req); err != nil {
			t.Fatalf("ModifyRequest(): got %v, want no error", err)
		}

		ids[idx] = logger.Entries.Entries()[idx].ID
	}

	for idx, url := range urls {
		if got, want := logger.Entries.RetrieveEntry(ids[idx]).Request.URL, url; got != want {
			t.Errorf("RetrieveEntry(): got %q, want %q", got, want)
		}
	}

	matcher := func(e *Entry) bool {
		return strings.Contains(e.Request.URL, "1.example.com")
	}

	if got, want := logger.Entries.RemoveEntry(ids[0]).Request.URL, urls[0]; got != want {
		t.Errorf("RemoveEntry: got %q, want %q", got, want)
	}

	if got := logger.Entries.RemoveEntry(ids[0]); got != nil {
		t.Errorf("RemoveEntry: should not have retrieve an entry")
	}

	if got, want := logger.Entries.RemoveMatches(matcher)[0].Request.URL, urls[1]; got != want {
		t.Errorf("RemoveMatches: got %q, want %q", got, want)
	}

	if got, want := logger.Entries.RetrieveEntry(ids[2]).Request.URL, urls[2]; got != want {
		t.Errorf("RemoveEntry got %q, want %q", got, want)
	}

	if got := logger.Entries.RetrieveEntry(""); got != nil {
		t.Errorf("RetrieveEntry: should not have retrieve an entry")
	}
}
