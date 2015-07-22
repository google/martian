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

// Package session provides contextual information about a single HTTP/S
// connection and its associated requests and responses.
package session

import (
	"errors"
	"sync"
)

// Context provides information and storage for a single request/response pair.
// Contexts are linked to shared session that is used for multiple requests on
// a single connection.
type Context struct {
	session       *session
	skipRoundTrip bool

	valmu sync.RWMutex
	vals  map[string]interface{}
}

type session struct {
	sync.Mutex
	id     string
	secure bool
}

// FromContext builds a new context from an existing context. The new context
// shares the same session as the passed context, but does not inherit any of
// its request specific values. The context cannot be nil.
func FromContext(ctx *Context) (*Context, error) {
	if ctx == nil {
		return nil, errors.New("session: cannot build context from nil")
	}

	return &Context{
		session: ctx.session,
		vals:    make(map[string]interface{}),
	}, nil
}

// NewContext builds a blank context.
func NewContext() *Context {
	return &Context{
		session: &session{},
		vals:    make(map[string]interface{}),
	}
}

// SetSessionID sets the ID for the session. The ID will be persisted across
// multiple requests and responses.
func (ctx *Context) SetSessionID(id string) {
	ctx.session.Lock()
	defer ctx.session.Unlock()

	ctx.session.id = id
}

// SessionID returns the session ID.
func (ctx *Context) SessionID() string {
	ctx.session.Lock()
	defer ctx.session.Unlock()

	return ctx.session.id
}

// IsSecure returns whether the current request is from a secure connection,
// such as when modifying a request from a TLS connection that has been MITM'd.
func (ctx *Context) IsSecure() bool {
	ctx.session.Lock()
	defer ctx.session.Unlock()

	return ctx.session.secure
}

// MarkSecure marks the session as secure.
func (ctx *Context) MarkSecure() {
	ctx.session.Lock()
	defer ctx.session.Unlock()

	ctx.session.secure = true
}

// Get takes key and returns the associated value from the context.
func (ctx *Context) Get(key string) (interface{}, bool) {
	ctx.valmu.RLock()
	val, ok := ctx.vals[key]
	ctx.valmu.RUnlock()

	return val, ok
}

// Set takes a key and associates it with val in the context. The value is
// persisted for the duration of the request and is removed on the following
// request.
func (ctx *Context) Set(key string, val interface{}) {
	ctx.valmu.Lock()
	defer ctx.valmu.Unlock()

	ctx.vals[key] = val
}

// SkipRoundTrip skips the round trip for the current request.
func (ctx *Context) SkipRoundTrip() {
	ctx.skipRoundTrip = true
}

// SkippingRoundTrip returns whether the current round trip will be skipped.
func (ctx *Context) SkippingRoundTrip() bool {
	return ctx.skipRoundTrip
}
