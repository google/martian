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
	"testing"

	"github.com/google/martian/parse"
)

func TestCookieModifier(t *testing.T) {
	mod := NewModifier()
	if mod == nil {
		t.Fatal("mod is nil")
	}
}

func TestModifierFromJSON(t *testing.T) {
	msg := []byte(`{
		"cache.Modifier": {
			"scope": ["request", "response"],
			"file": "/dev/null",
			"bucket": "martian"
		}
	}`)

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got %v, want no error", err)
	}

	// debug
	if r == nil {
		t.Fatal("result is nil")
	}
}
