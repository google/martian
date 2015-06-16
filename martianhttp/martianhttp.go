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

// Package martianhttp provides HTTP handlers for managing the state of a martian.Proxy.
package martianhttp

import (
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

// Modifier is a locking modifier that is configured via http.Handler.
type Modifier struct {
	mu     sync.RWMutex
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier
}

// NewModifier returns a new martianhttp.Modifier.
func NewModifier() *Modifier {
	return &Modifier{}
}

// ModifyRequest runs reqmod.
func (m *Modifier) ModifyRequest(ctx *martian.Context, req *http.Request) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.reqmod == nil {
		return nil
	}

	return m.reqmod.ModifyRequest(ctx, req)
}

// ModifyResponse runs resmod.
func (m *Modifier) ModifyResponse(ctx *martian.Context, res *http.Response) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.resmod == nil {
		return nil
	}

	return m.resmod.ModifyResponse(ctx, res)
}

// VerifyRequests verifies reqmod, iff reqmod is a RequestVerifier.
func (m *Modifier) VerifyRequests() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if reqv, ok := m.reqmod.(verify.RequestVerifier); ok {
		return reqv.VerifyRequests()
	}

	return nil
}

// VerifyResponses verifies resmod, iff resmod is a ResponseVerifier.
func (m *Modifier) VerifyResponses() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if resv, ok := m.resmod.(verify.ResponseVerifier); ok {
		return resv.VerifyResponses()
	}

	return nil
}

// ResetRequestVerifications resets verifications on reqmod, iff reqmod is a
// RequestVerifier.
func (m *Modifier) ResetRequestVerifications() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if reqv, ok := m.reqmod.(verify.RequestVerifier); ok {
		reqv.ResetRequestVerifications()
	}
}

// ResetResponseVerifications resets verifications on resdmo, iff resmod is a
// ResponseVerifier.
func (m *Modifier) ResetResponseVerifications() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if resv, ok := m.resmod.(verify.ResponseVerifier); ok {
		resv.ResetResponseVerifications()
	}
}

// ServeHTTP accepts a POST request with a body containing a modifier as a JSON
// message and updates the contained reqmod and resmod with the parsed
// modifier.
func (m *Modifier) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.Header().Set("Allow", "POST")
		rw.WriteHeader(405)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		martian.Errorf("error reading request body: %v", err)
		return
	}
	req.Body.Close()

	r, err := parse.FromJSON(body)
	if err != nil {
		http.Error(rw, err.Error(), 400)
		martian.Errorf("error parsing JSON: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.reqmod = r.RequestModifier()
	m.resmod = r.ResponseModifier()
}
