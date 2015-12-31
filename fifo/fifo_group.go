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

package fifo

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

// Group is a martian.RequestResponseModifier that maintains lists of
// request and response modifiers executed on a first-in, first-out basis.
type Group struct {
	reqmu   sync.RWMutex
	reqmods []martian.RequestModifier

	resmu   sync.RWMutex
	resmods []martian.ResponseModifier
}

type groupJSON struct {
	Modifiers []json.RawMessage    `json:"modifiers"`
	Scope     []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("fifo.Group", groupFromJSON)
}

// NewGroup returns a modifier group.
func NewGroup() *Group {
	return &Group{}
}

// AddRequestModifier adds a RequestModifier to the group's list of request modifiers.
func (g *Group) AddRequestModifier(reqmod martian.RequestModifier) {
	g.reqmu.Lock()
	defer g.reqmu.Unlock()

	g.reqmods = append(g.reqmods, reqmod)
}

// AddResponseModifier adds a ResponseModifier to the group's list of response modifiers.
func (g *Group) AddResponseModifier(resmod martian.ResponseModifier) {
	g.resmu.Lock()
	defer g.resmu.Unlock()

	g.resmods = append(g.resmods, resmod)
}

// ModifyRequest modifies the request. If an error is returned by a
// RequestModifier the error is returned and no further modifiers are run.
func (g *Group) ModifyRequest(req *http.Request) error {
	g.reqmu.RLock()
	defer g.reqmu.RUnlock()

	for _, reqmod := range g.reqmods {
		if err := reqmod.ModifyRequest(req); err != nil {
			return err
		}
	}

	return nil
}

// ModifyResponse modifies the request. If an error is returned by a
// ResponseModifier the error is returned and no further modifiers are run.
func (g *Group) ModifyResponse(res *http.Response) error {
	g.resmu.RLock()
	defer g.resmu.RUnlock()

	for _, resmod := range g.resmods {
		if err := resmod.ModifyResponse(res); err != nil {
			return err
		}
	}

	return nil
}

// groupFromJSON builds a fifo.Group from JSON.
//
// Example JSON:
// {
//   "fifo.Group" : {
//     "scope": ["request", "result"],
//     "modifiers": [
//       { ... },
//       { ... },
//     ]
//   }
// }
func groupFromJSON(b []byte) (*parse.Result, error) {
	msg := &groupJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	g := NewGroup()

	for _, m := range msg.Modifiers {
		r, err := parse.FromJSON(m)
		if err != nil {
			return nil, err
		}

		reqmod := r.RequestModifier()
		if reqmod != nil {
			g.AddRequestModifier(reqmod)
		}

		resmod := r.ResponseModifier()
		if resmod != nil {
			g.AddResponseModifier(resmod)
		}
	}

	return parse.NewResult(g, msg.Scope)
}
