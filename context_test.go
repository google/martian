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

package martian

import (
	"errors"
	"testing"
)

func TestAuthReset(t *testing.T) {
	auth := &Auth{
		ID:	"username",
		Error:	errors.New("invalid auth"),
	}

	auth.Reset()

	if got, want := auth.ID, ""; got != want {
		t.Errorf("auth.ID: got %q, want %q", got, want)
	}
	if err := auth.Error; err != nil {
		t.Errorf("auth.Error: got %v, want no error", err)
	}
}
