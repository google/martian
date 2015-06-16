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

package parse

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/martian"
)

func TestFromJSON(t *testing.T) {
	wasRun := false

	Register("parse.Key", func(b []byte) (*Result, error) {
		m := martian.RequestModifierFunc(
			func(*martian.Context, *http.Request) error {
				wasRun = true
				return nil
			})

		msg := &struct {
			Scope []ModifierType `json:"scope"`
		}{}

		if err := json.Unmarshal(b, msg); err != nil {
			return nil, err
		}

		return NewResult(m, msg.Scope)
	})

	rawMsg := []byte(`{
	  "parse.Key": {
      "scope":["request"]
    }
	}`)

	r, err := FromJSON(rawMsg)
	if err != nil {
		t.Fatalf("FromJSON(): got %v, want no error", err)
	}

	reqmod := r.RequestModifier()
	if reqmod == nil {
		t.Fatal("FromJSON(): got nil, want not nil")
	}

	resmod := r.ResponseModifier()
	if resmod != nil {
		t.Fatal("FromJSON(): got nil, want not nil")
	}

	err = reqmod.ModifyRequest(nil, nil)
	if err != nil {
		t.Fatalf("reqmod.ModifyRequest(): got %v, want no error", err)
	}
	if !wasRun {
		t.Error("FromJSON(): got false, want true")
	}

}

func TestResultRequestResponseModifierCorrectScope(t *testing.T) {
	mod := struct {
		martian.RequestModifier
		martian.ResponseModifier
	}{
		RequestModifier: martian.RequestModifierFunc(
			func(*martian.Context, *http.Request) error {
				return nil
			}),
		ResponseModifier: martian.ResponseModifierFunc(
			func(*martian.Context, *http.Response) error {
				return nil
			}),
	}
	result := &Result{
		reqmod: mod,
		resmod: nil,
	}
	reqmod := result.RequestModifier()
	if reqmod == nil {
		t.Error("result.RequestModifier: got nil, want not nil")
	}

	resmod := result.ResponseModifier()
	if resmod != nil {
		t.Errorf("result.ResponseModifier: got %v, want nil", resmod)
	}

	result = &Result{
		reqmod: nil,
		resmod: mod,
	}
	reqmod = result.RequestModifier()
	if reqmod != nil {
		t.Errorf("result.RequestModifier: got %v, want nil", reqmod)
	}

	resmod = result.ResponseModifier()
	if resmod == nil {
		t.Error("result.ResponseModifier: got nil, want not nil")
	}
}

func TestParseUnknownModifierReturnsError(t *testing.T) {
	rawMsg := `
	{
	  "unknown.Key": {
      "scope": ["request", "response"]
		}
	}`

	_, err := FromJSON([]byte(rawMsg))
	if _, ok := err.(ErrUnknownModifier); !ok {
		t.Fatalf("FromJSON(): got %v, want ErrUnknownModifier", err)
	}
}
