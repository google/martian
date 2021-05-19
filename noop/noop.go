// Copyright 2021 Google Inc. All rights reserved.
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

// Package noop provides a martian.RequestResponseModifier that does not
// modify the request or response.
package noop

import (
	"encoding/json"
	"net/http"

	"google3/third_party/golang/martian/log/log"
	"google3/third_party/golang/martian/martian"
	"google3/third_party/golang/martian/parse/parse"
)

func init() {
	parse.Register("noop.Modifier", modifierFromJSON)
}

type noopModifier struct {
	id string
}

// Noop returns a modifier that does not change the request or the response.
func Noop(id string) martian.RequestResponseModifier {
	return &noopModifier{
		id: id,
	}
}

// ModifyRequest logs a debug line.
func (nm *noopModifier) ModifyRequest(*http.Request) error {
	log.Debugf("noopModifier: %s: no request modification applied", nm.id)
	return nil
}

// ModifyResponse logs a debug line.
func (nm *noopModifier) ModifyResponse(*http.Response) error {
	log.Debugf("noopModifier: %s: no response modification applied", nm.id)
	return nil
}

type modifierJSON struct {
	Name  string               `json:"name"`
	Scope []parse.ModifierType `json:"scope"`
}

// modifierFromJSON takes a JSON message as a byte slice and returns
// a headerModifier and an error.
//
// Example JSON configuration message:
// {
//  "scope": ["request", "result"],
//  "name": "noop-name",
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	modifier := Noop(msg.Name)

	return parse.NewResult(modifier, msg.Scope)
}

