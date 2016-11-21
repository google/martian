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

package static

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
)

func TestStaticModifierOnRequest(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "static_file_modifier_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir(): got %v, want no error", err)
	}

	if err := ioutil.WriteFile(path.Join(tmpdir, "sfmtest"), []byte("test file"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(): got %v, want no error", err)
	}

	req, err := http.NewRequest("GET", "/sfmtest", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}

	_, remove, err := martian.TestContext(req, nil, nil)
	if err != nil {
		t.Fatalf("TestContext(): got %v, want no error", err)
	}
	defer remove()

	res := proxyutil.NewResponse(http.StatusOK, nil, req)

	mod := NewModifier(tmpdir)
	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)

	}
	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	res.Body.Close()

	if want := []byte("test file"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}
