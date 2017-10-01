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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/martian/parse"
)

func TestCacheModifier(t *testing.T) {
	f, err := ioutil.TempFile("", "cache_test")
	if err != nil {
		t.Fatalf("ioutil.TempFile(): got error %v, want no error", err)
	}
	defer os.RemoveAll(f.Name())
	mod, err := NewModifier(f.Name(), "foo", true, true)
	if err != nil {
		t.Fatalf("NewModifier: got error %v, want no error", err)
	}
	if mod == nil {
		t.Fatal("NewModifier: mod is nil")
	}
}

func TestModifierFromJSON(t *testing.T) {
	f, err := ioutil.TempFile("", "cache_test")
	if err != nil {
		t.Fatalf("ioutil.TempFile(): got error %v, want no error", err)
	}
	defer os.RemoveAll(f.Name())
	msg := []byte(fmt.Sprintf(`{
		"cache.Modifier": {
			"scope": ["request", "response"],
			"file": "%s",
			"bucket": "foo",
			"replay": true,
			"update": true
		}
	}`, f.Name()))

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got error %v, want no error", err)
	}
	if r == nil {
		t.Fatal("parse.FromJSON(): result is nil")
	}
}
