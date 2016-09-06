// Copyright 2016 Google Inc. All rights reserved.
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

package api

import (
	"net/http"
	"testing"

	"github.com/google/martian"
)

func TestApiForwarder(t *testing.T) {
	forwarder := NewForwarder("apihost.com")

	req, err := http.NewRequest("GET", "https://localhost:8080/configure", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req, nil, nil)
	if err != nil {
		t.Fatalf("TestContext(): got %v, want no error", err)
	}
	defer remove()

	if err := forwarder.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got, want := req.URL.Scheme, "http"; got != want {
		t.Errorf("req.URL.Scheme: got %s, want %s", got, want)
	}
	if got, want := req.URL.Host, "apihost.com"; got != want {
		t.Errorf("req.URL.Host: got %s, want %s", got, want)
	}

	if !ctx.SkippingLogging() {
		t.Errorf("SkippingLogging: got false, want true")
	}
}
