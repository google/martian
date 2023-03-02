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
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Context provides information and storage for a single request/response pair.
// Contexts are linked to shared session that is used for multiple requests on
// a single connection.
type Context struct {
	session *Session
	id      uint64

	mu            sync.RWMutex
	vals          map[string]any
	skipRoundTrip bool
	skipLogging   bool
	apiRequest    bool
}

// Session provides information and storage about a connection.
type Session struct {
	mu       sync.RWMutex
	secure   bool
	hijacked bool
	conn     net.Conn
	brw      *bufio.ReadWriter
	vals     map[string]any
}

const marianKey string = "martian.Context"

// NewContext returns a context for the in-flight HTTP request.
func NewContext(req *http.Request) *Context {
	v := req.Context().Value(marianKey)
	if v == nil {
		return nil
	}
	return v.(*Context)
}

// TestContext builds a new session and associated context and returns the
// context and a function to remove the associated context. If it fails to
// generate either a new session or a new context it will return an error.
// Intended for tests only.
func TestContext(req *http.Request, conn net.Conn, bw *bufio.ReadWriter) (ctx *Context, remove func(), err error) {
	nop := func() {}

	ctx = NewContext(req)
	if ctx != nil {
		return ctx, nop, nil
	}

	ctx = withSession(newSession(conn, bw))
	*req = *req.Clone(ctx.addToContext(req.Context()))

	return ctx, nop, nil
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

// MarkInsecure marks the session as insecure.
func (s *Session) MarkInsecure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.secure = false
}

// Hijack takes control of the connection from the proxy. No further action
// will be taken by the proxy and the connection will be closed following the
// return of the hijacker.
func (s *Session) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hijacked {
		return nil, nil, fmt.Errorf("martian: session has already been hijacked")
	}
	s.hijacked = true

	return s.conn, s.brw, nil
}

// Hijacked returns whether the connection has been hijacked.
func (s *Session) Hijacked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.hijacked
}

// setConn resets the underlying connection and bufio.ReadWriter of the
// session. Used by the proxy when the connection is upgraded to TLS.
func (s *Session) setConn(conn net.Conn, brw *bufio.ReadWriter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.conn = conn
	s.brw = brw
}

// Get takes key and returns the associated value from the session.
func (s *Session) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.vals[key]

	return val, ok
}

// Set takes a key and associates it with val in the session. The value is
// persisted for the entire session across multiple requests and responses.
func (s *Session) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vals == nil {
		s.vals = make(map[string]any)
	}

	s.vals[key] = val
}

// addToContext returns context.Context with the current context to the passed context.
func (ctx *Context) addToContext(rctx context.Context) context.Context {
	if rctx == nil {
		rctx = context.Background()
	}
	return context.WithValue(rctx, marianKey, ctx)
}

// Session returns the session for the context.
func (ctx *Context) Session() *Session {
	return ctx.session
}

// ID returns the context ID.
func (ctx *Context) ID() string {
	return strconv.FormatUint(ctx.id, 16)
}

// Get takes key and returns the associated value from the context.
func (ctx *Context) Get(key string) (any, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	val, ok := ctx.vals[key]

	return val, ok
}

// Set takes a key and associates it with val in the context. The value is
// persisted for the duration of the request and is removed on the following
// request.
func (ctx *Context) Set(key string, val any) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.vals == nil {
		ctx.vals = make(map[string]any)
	}

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

// SkipLogging skips logging by Martian loggers for the current request.
func (ctx *Context) SkipLogging() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.skipLogging = true
}

// SkippingLogging returns whether the current request / response pair will be logged.
func (ctx *Context) SkippingLogging() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.skipLogging
}

// APIRequest marks the requests as a request to the proxy API.
func (ctx *Context) APIRequest() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.apiRequest = true
}

// IsAPIRequest returns true when the request patterns matches a pattern in the proxy
// mux. The mux is usually defined as a parameter to the api.Forwarder, which uses
// http.DefaultServeMux by default.
func (ctx *Context) IsAPIRequest() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.apiRequest
}

// newSession builds a new session.
func newSession(conn net.Conn, brw *bufio.ReadWriter) *Session {
	return &Session{
		conn: conn,
		brw:  brw,
	}
}

var nextID atomic.Uint64

func init() {
	nextID.Store(uint64(time.Now().UnixMilli()))
}

// withSession builds a new context from an existing session.
// Session must be non-nil.
func withSession(s *Session) *Context {
	return &Context{
		session: s,
		id:      nextID.Add(1),
	}
}
