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

package auth

import (
	"net/http"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
)

func TestEmptyIDReturnsError(t *testing.T) {
	f := NewFilter()

	if err := f.SetRequestModifier("", nil); err != ErrIDRequired {
		t.Errorf("SetRequestModifier(): got %v, want ErrIDRequired", err)
	}

	if err := f.SetResponseModifier("", nil); err != ErrIDRequired {
		t.Errorf("SetResponseModifier(): got %v, want ErrIDRequired", err)
	}
}

func TestFilter(t *testing.T) {
	f := NewFilter()
	if reqmod := f.RequestModifier("id"); reqmod != nil {
		t.Fatalf("f.RequestModifier(%q): got reqmod, want no reqmod", "id")
	}
	if resmod := f.ResponseModifier("id"); resmod != nil {
		t.Fatalf("f.ResponseModifier(%q): got resmod, want no resmod", "id")
	}

	f.SetRequestModifier("id", martian.RequestModifierFunc(
		func(*martian.Context, *http.Request) error {
			return nil
		}))

	f.SetResponseModifier("id", martian.ResponseModifierFunc(
		func(*martian.Context, *http.Response) error {
			return nil
		}))

	if reqmod := f.RequestModifier("id"); reqmod == nil {
		t.Errorf("f.RequestModifier(%q): got no reqmod, want reqmod", "id")
	}
	if resmod := f.ResponseModifier("id"); resmod == nil {
		t.Errorf("f.ResponseModifier(%q): got no resmod, want resmod", "id")
	}

	f.SetRequestModifier("id", nil)
	f.SetResponseModifier("id", nil)
	if reqmod := f.RequestModifier("id"); reqmod != nil {
		t.Fatalf("f.RequestModifier(%q): got reqmod, want no reqmod", "id")
	}
	if resmod := f.ResponseModifier("id"); resmod != nil {
		t.Fatalf("f.ResponseModifier(%q): got resmod, want no resmod", "id")
	}
}

func TestModifyRequest(t *testing.T) {
	f := NewFilter()

	modifierRun := false
	f.SetRequestModifier("id", martian.RequestModifierFunc(
		func(*martian.Context, *http.Request) error {
			modifierRun = true
			return nil
		}))

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}
	ctx := martian.NewContext()

	// No ID, auth required.
	f.SetAuthRequired(true)

	if err := f.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if ctx.Auth.Error == nil {
		t.Error("ctx.Auth.Error: got nil, want error")
	}
	if modifierRun {
		t.Error("modifierRun: got true, want false")
	}

	// No ID, auth not required.
	f.SetAuthRequired(false)
	ctx.Auth.Error = nil

	if err := f.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if ctx.Auth.Error != nil {
		t.Errorf("ctx.Auth.Error: got %v, want no error", err)
	}
	if modifierRun {
		t.Error("modifierRun: got true, want false")
	}

	// Valid ID.
	ctx.Auth.ID = "id"
	ctx.Auth.Error = nil
	if err := f.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if ctx.Auth.Error != nil {
		t.Errorf("ctx.Auth.Error: got %v, want no error", ctx.Auth.Error)
	}
	if !modifierRun {
		t.Error("modifierRun: got false, want true")
	}
}

func TestModifyResponse(t *testing.T) {
	f := NewFilter()

	modifierRun := false
	f.SetResponseModifier("id", martian.ResponseModifierFunc(
		func(*martian.Context, *http.Response) error {
			modifierRun = true
			return nil
		}))

	res := proxyutil.NewResponse(200, nil, nil)
	ctx := martian.NewContext()

	// No ID, auth required.
	f.SetAuthRequired(true)

	if err := f.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if ctx.Auth.Error == nil {
		t.Error("ctx.Auth.Error: got nil, want error")
	}
	if modifierRun {
		t.Error("modifierRun: got true, want false")
	}

	// No ID, no auth required.
	f.SetAuthRequired(false)
	ctx.Auth.Error = nil

	if err := f.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if ctx.Auth.Error != nil {
		t.Errorf("ctx.Auth.Error: got %v, want no error", ctx.Auth.Error)
	}
	if modifierRun {
		t.Error("modifierRun: got true, want false")
	}

	// Valid ID.
	ctx.Auth.ID = "id"
	ctx.Auth.Error = nil

	if err := f.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if ctx.Auth.Error != nil {
		t.Errorf("ctx.Auth.Error: got %v, want no error", ctx.Auth.Error)
	}
	if !modifierRun {
		t.Error("modifierRun: got false, want true")
	}
}
