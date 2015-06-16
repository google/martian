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

// Package auth provides filtering support for a martian.Proxy based on ctx.Auth.ID.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/martian"
)

// Filter filters RequestModifiers and ResponseModifiers by ctx.Auth.ID.
type Filter struct {
	authRequired bool

	mu      sync.RWMutex
	reqmods map[string]martian.RequestModifier
	resmods map[string]martian.ResponseModifier
}

// ErrIDRequired indicates that the filter must have an ID.
var ErrIDRequired = errors.New("ID required")

// NewFilter returns a new auth.Filter.
func NewFilter() *Filter {
	return &Filter{
		reqmods: make(map[string]martian.RequestModifier),
		resmods: make(map[string]martian.ResponseModifier),
	}
}

// SetAuthRequired determines whether the ctx.Auth.ID must have an associated
// RequestModifier or ResponseModifier. If true, it will set ctx.Auth.Error.
func (f *Filter) SetAuthRequired(required bool) {
	f.authRequired = required
}

// SetRequestModifier sets the RequestModifier for the given ID. It will
// overwrite any existing modifier with the same ID.
// Returns ErrIDRequired if id is empty.
func (f *Filter) SetRequestModifier(id string, reqmod martian.RequestModifier) error {
	if id == "" {
		return ErrIDRequired
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if reqmod != nil {
		f.reqmods[id] = reqmod
	} else {
		delete(f.reqmods, id)
	}

	return nil
}

// SetResponseModifier sets the ResponseModifier for the given ID. It will
// overwrite any existing modifier with the same ID.
// Returns ErrIDRequired if id is empty.
func (f *Filter) SetResponseModifier(id string, resmod martian.ResponseModifier) error {
	if id == "" {
		return ErrIDRequired
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if resmod != nil {
		f.resmods[id] = resmod
	} else {
		delete(f.resmods, id)
	}

	return nil
}

// RequestModifier retrieves the RequestModifier for the given ID. Returns nil
// if no modifier exists for the given ID.
func (f *Filter) RequestModifier(id string) martian.RequestModifier {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.reqmods[id]
}

// ResponseModifier retrieves the ResponseModifier for the given ID. Returns nil
// if no modifier exists for the given ID.
func (f *Filter) ResponseModifier(id string) martian.ResponseModifier {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.resmods[id]
}

// ModifyRequest runs the RequestModifier for the associated ctx.Auth.ID. If no
// modifier is found for ctx.Auth.ID then ctx.Auth.Error is set.
func (f *Filter) ModifyRequest(ctx *martian.Context, req *http.Request) error {
	if reqmod := f.reqmods[ctx.Auth.ID]; reqmod != nil {
		return reqmod.ModifyRequest(ctx, req)
	}

	return f.requireKnownAuth(ctx)
}

// ModifyResponse runs the ResponseModifier for the associated ctx.Auth.ID. If
// no modifier is found for ctx.Auth.ID then ctx.Auth.Error is set.
func (f *Filter) ModifyResponse(ctx *martian.Context, res *http.Response) error {
	if resmod := f.resmods[ctx.Auth.ID]; resmod != nil {
		return resmod.ModifyResponse(ctx, res)
	}

	return f.requireKnownAuth(ctx)
}

func (f *Filter) requireKnownAuth(ctx *martian.Context) error {
	_, reqok := f.reqmods[ctx.Auth.ID]
	_, resok := f.resmods[ctx.Auth.ID]

	if !reqok && !resok && f.authRequired {
		ctx.Auth.Error = fmt.Errorf("no modifiers found for %s", ctx.Auth.ID)
	}

	return nil
}
