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

package session

import "testing"

func TestContext(t *testing.T) {
	if _, err := FromContext(nil); err == nil {
		t.Error("FromContext(nil): got nil, want error")
	}

	ctx := NewContext()

	ctx.SetSessionID("id")
	if got, want := ctx.SessionID(), "id"; got != want {
		t.Errorf("ctx.SessionID(): got %q, want %q", got, want)
	}

	ctx.MarkSecure()
	if !ctx.IsSecure() {
		t.Error("ctx.IsSecure(): got false, want true")
	}

	ctx.Set("key", "value")
	got, ok := ctx.Get("key")
	if !ok {
		t.Errorf("ctx.Get(%q): got !ok, want ok", "key")
	}
	if want := "value"; got != want {
		t.Errorf("ctx.Get(%q): got %q, want %q", "key", got, want)
	}

	ctx.SkipRoundTrip()
	if !ctx.SkippingRoundTrip() {
		t.Error("ctx.SkippingRoundTrip(): got false, want true")
	}

	ctx2, err := FromContext(ctx)
	if err != nil {
		t.Fatalf("FromContext(ctx): got %v, want no error", err)
	}

	if got, want := ctx2.SessionID(), "id"; got != want {
		t.Errorf("ctx2.SessionID(): got %q, want %q", got, want)
	}
	if !ctx2.IsSecure() {
		t.Error("ctx2.IsSecure(): got false, want true")
	}

	if ctx2.SkippingRoundTrip() {
		t.Error("ctx2.SkippingRoundTrip(): got true, want false")
	}
	if _, ok := ctx2.Get("key"); ok {
		t.Errorf("ctx2.Get(%q): got ok, want !ok", "key")
	}
}
