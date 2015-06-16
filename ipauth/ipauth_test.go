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

package ipauth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
)

func TestModifyRequest(t *testing.T) {
	m := NewModifier()
	ctx := martian.NewContext()

	req, err := http.NewRequest("CONNECT", "https://www.example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	if err := m.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if got, want := ctx.Auth.ID, ""; got != want {
		t.Errorf("ctx.Auth.ID: got %q, want %q", got, want)
	}

	// IP with port and modifier with error.
	req.RemoteAddr = "1.1.1.1:8111"
	reqErr := errors.New("request modifier failure")
	m.SetRequestModifier(martian.RequestModifierFunc(
		func(*martian.Context, *http.Request) error {
			return reqErr
		}))

	if err := m.ModifyRequest(ctx, req); err != reqErr {
		t.Fatalf("ModifyConnectRequest(): got %v, want %v", err, reqErr)
	}
	if got, want := ctx.Auth.ID, "1.1.1.1"; got != want {
		t.Errorf("ctx.Auth.ID: got %q, want %q", got, want)
	}

	// IP without port and modifier with auth error.
	req.RemoteAddr = "4.4.4.4"
	m.SetRequestModifier(martian.RequestModifierFunc(
		func(ctx *martian.Context, req *http.Request) error {
			ctx.Auth.Error = errors.New("auth error")
			return nil
		}))

	if err := m.ModifyRequest(ctx, req); err != nil {
		t.Fatalf("ModifyConnectRequest(): got %v, want no error", err)
	}
	if got, want := ctx.Auth.ID, "4.4.4.4"; got != want {
		t.Errorf("ctx.Auth.ID: got %q, want %q", got, want)
	}
	if !ctx.SkipRoundTrip {
		t.Error("ctx.SkipRoundTrip: got false, want true")
	}
}

func TestModifyResponse(t *testing.T) {
	m := NewModifier()
	ctx := martian.NewContext()

	res := proxyutil.NewResponse(200, nil, nil)
	if err := m.ModifyResponse(ctx, res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	// Modifier with error.
	resErr := errors.New("response modification failure")
	m.SetResponseModifier(martian.ResponseModifierFunc(
		func(*martian.Context, *http.Response) error {
			return resErr
		}))
	if err := m.ModifyResponse(ctx, res); err != resErr {
		t.Fatalf("ModifyResponse(): got %v, want %v", err, resErr)
	}

	// Modifier with auth error.
	authErr := errors.New("auth error")
	m.SetResponseModifier(martian.ResponseModifierFunc(
		func(ctx *martian.Context, res *http.Response) error {
			ctx.Auth.Error = authErr
			return nil
		}))

	ctx.Auth.ID = "bad-auth"
	ctx.SkipRoundTrip = true
	if err := m.ModifyResponse(ctx, res); err != authErr {
		t.Fatalf("ModifyResponse(): got %v, want %v", err, authErr)
	}

	// ctx.Auth should be reset.
	if got, want := ctx.Auth.ID, ""; got != want {
		t.Errorf("ctx.Auth.ID: got %q, want %q", got, want)
	}
	if ctx.SkipRoundTrip {
		t.Error("ctx.SkipRoundTrip: got true, want false")
	}
}
