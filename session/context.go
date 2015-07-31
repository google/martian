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

import "sync"

// Context provides information and storage for a single request/response pair.
// Contexts are linked to shared session that is used for multiple requests on
// a single connection.
type Context struct {
	session *Session

	mu            sync.RWMutex
	vals          map[string]interface{}
	skipRoundTrip bool
}

// Session provides information and storage about a connection.
type Session struct {
	mu     sync.RWMutex
	id     string
	secure bool
	vals   map[string]interface{}
}

// FromContext builds a new context from an existing context. The new context
// shares the same session as the passed context, but does not inherit any of
// its request specific values. If ctx is nil, a new context and session are
// created.
func FromContext(ctx *Context) *Context {
	session := &Session{
		vals: make(map[string]interface{}),
	}

	if ctx != nil {
		session = ctx.session
	}

	return &Context{
		session: session,
		vals:    make(map[string]interface{}),
	}
}

// SetID sets the ID for the session. The ID will be persisted across
// multiple requests and responses.
func (s *Session) SetID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.id = id
}

// ID returns the session ID.
func (s *Session) ID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.id
}

// IsSecure returns whether the current session is from a secure connection,
// such as when receiving requests from a TLS connection that has been MITM'd.
func (s *Session) IsSecure() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.secure
}

// MarkSecure marks the session as secure.
func (s *Session) MarkSecure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.secure = true
}

// Get takes key and returns the associated value from the session.
func (s *Session) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.vals[key]

	return val, ok
}

// Set takes a key and associates it with val in the session. The value is
// persisted for the entire session across multiple requests and responses.
func (s *Session) Set(key string, val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vals[key] = val
}

// GetSession returns the session for the context.
func (ctx *Context) GetSession() *Session {
	return ctx.session
}

// Get takes key and returns the associated value from the context.
func (ctx *Context) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	val, ok := ctx.vals[key]

	return val, ok
}

// Set takes a key and associates it with val in the context. The value is
// persisted for the duration of the request and is removed on the following
// request.
func (ctx *Context) Set(key string, val interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.vals[key] = val
}

// SkipRoundTrip skips the round trip for the current request.
func (ctx *Context) SkipRoundTrip() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.skipRoundTrip = true
}

// SkippingRoundTrip returns whether the current round trip will be skipped.
func (ctx *Context) SkippingRoundTrip() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.skipRoundTrip
}
