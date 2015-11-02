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
	ctx, err := FromContext(nil)
	if err != nil {
		t.Fatalf("FromContext(nil): got %v, want no error", err)
	}

	if len(ctx.ID()) != 16 {
		t.Errorf("ctx.ID(): got %q, want 16 character random ID", ctx.ID())
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

	session := ctx.Session()

	if len(session.ID()) != 16 {
		t.Errorf("session.ID(): got %q, want 16 character random ID", session.ID())
	}

	session.MarkSecure()
	if !session.IsSecure() {
		t.Error("session.IsSecure(): got false, want true")
	}

	session.Set("key", "value")
	got, ok = session.Get("key")
	if !ok {
		t.Errorf("session.Get(%q): got !ok, want ok", "key")
	}
	if want := "value"; got != want {
		t.Errorf("session.Get(%q): got %q, want %q", "key", got, want)
	}

	ctx2, err := FromContext(ctx)
	if err != nil {
		t.Fatalf("FromContext(ctx): got %v, want no error", err)
	}

	if ctx2.SkippingRoundTrip() {
		t.Error("ctx2.SkippingRoundTrip(): got true, want false")
	}
	if _, ok := ctx2.Get("key"); ok {
		t.Errorf("ctx2.Get(%q): got ok, want !ok", "key")
	}

	session2 := ctx2.Session()
	if got, want := session2.ID(), session.ID(); got != want {
		t.Errorf("session2.ID(): got %q, want %q", got, want)
	}

	if !session2.IsSecure() {
		t.Error("session2.IsSecure(): got false, want true")
	}

	got, ok = session2.Get("key")
	if !ok {
		t.Errorf("session2.Get(%q): got !ok, want ok", "key")
	}
	if want := "value"; got != want {
		t.Errorf("session2.Get(%q): got %q, want %q", "key", got, want)
	}
}
