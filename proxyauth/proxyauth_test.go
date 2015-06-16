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

package proxyauth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
)

func encode(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

func TestModifyRequest(t *testing.T) {
	m := NewModifier()
	ctx := martian.NewContext()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Proxy-Authorization", "Basic "+encode("user:pass"))

	if err := m.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := ctx.Auth.ID, "user:pass"; got != want {
		t.Fatalf("ctx.Auth.ID: got %q, want %q", got, want)
	}

	modifierRun := false
	m.SetRequestModifier(martian.RequestModifierFunc(
		func(*martian.Context, *http.Request) error {
			modifierRun = true
			return nil
		}))

	if err := m.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if !modifierRun {
		t.Error("modifierRun: got false, want true")
	}
}

func TestModifyResponse(t *testing.T) {
	m := NewModifier()
	ctx := martian.NewContext()
	res := proxyutil.NewResponse(200, nil, nil)

	if err := m.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	m.SetResponseModifier(martian.ResponseModifierFunc(
		func(*martian.Context, *http.Response) error {
			ctx.Auth.Error = errors.New("auth is required")
			return nil
		}))

	if err := m.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.StatusCode, http.StatusProxyAuthRequired; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	if got, want := res.Header.Get("Proxy-Authenticate"), "Basic"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Proxy-Authenticate", got, want)
	}
}

func TestModifyResponseResetAuth(t *testing.T) {
	auth := NewModifier()

	auth.SetRequestModifier(martian.RequestModifierFunc(
		func(ctx *martian.Context, req *http.Request) error {
			if ctx.Auth.ID != "secret:pass" {
				ctx.Auth.Error = errors.New("invalid auth")
			}
			return nil
		}))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Proxy-Authorization", "Basic "+encode("wrong:pass"))

	ctx := martian.NewContext()
	// This will set ctx.Auth.Error since the ID isn't "secret:pass".
	if err := auth.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	res := proxyutil.NewResponse(200, nil, req)
	if err := auth.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.StatusCode, 407; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}

	if got, want := ctx.Auth.ID, ""; got != want {
		t.Errorf("ctx.Auth.ID: got %q, want %q", got, want)
	}
	if err := ctx.Auth.Error; err != nil {
		t.Errorf("ctx.Auth.Error: got %v, want no error", err)
	}

	// This will be successful because the ID is "secret:pass".
	req.Header.Set("Proxy-Authorization", "Basic "+encode("secret:pass"))
	if err := auth.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	// Reset the response.
	res = proxyutil.NewResponse(200, nil, req)
	if err := auth.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
}
