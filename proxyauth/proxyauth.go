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

// Package proxyauth provides authentication support via the
// Proxy-Authorization header.
package proxyauth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/google/martian"
)

// Modifier is the proxy authentication modifier.
type Modifier struct {
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier
}

// NewModifier returns a new proxy authentication modifier.
func NewModifier() *Modifier {
	return &Modifier{}
}

// SetRequestModifier sets the request modifier.
func (m *Modifier) SetRequestModifier(reqmod martian.RequestModifier) {
	m.reqmod = reqmod
}

// SetResponseModifier sets the response modifier.
func (m *Modifier) SetResponseModifier(resmod martian.ResponseModifier) {
	m.resmod = resmod
}

// ModifyRequest sets the auth ID in the context from the request iff it has
// not already been set and runs reqmod.ModifyRequest. If the underlying
// modifier has indicated via ctx.Auth.Error that no valid auth credentials
// have been found we set ctx.SkipRoundTrip.
func (m *Modifier) ModifyRequest(ctx *martian.Context, req *http.Request) error {
	ctx.Auth.ID = id(req.Header)

	if m.reqmod == nil {
		return nil
	}

	if err := m.reqmod.ModifyRequest(ctx, req); err != nil {
		return err
	}

	if ctx.Auth.Error != nil {
		ctx.SkipRoundTrip = true
	}

	return nil
}

// ModifyResponse runs resmod.ModifyResponse and modifies the response to
// include the correct status code and headers if ctx.Auth.Error is present.
//
// If an error is returned from resmod.ModifyResponse it is returned.
func (m *Modifier) ModifyResponse(ctx *martian.Context, res *http.Response) error {
	var err error

	if m.resmod != nil {
		err = m.resmod.ModifyResponse(ctx, res)
	}

	if ctx.Auth.Error != nil {
		// Reset the auth so we don't get stuck in a failed auth loop
		// when dealing with Keep-Alive.
		ctx.Auth.Reset()
		ctx.SkipRoundTrip = false

		res.StatusCode = http.StatusProxyAuthRequired
		res.Header.Set("Proxy-Authenticate", "Basic")
	}

	return err
}

// id returns an ID derived from the Proxy-Authorization header username and password.
func id(header http.Header) string {
	id := strings.TrimPrefix(header.Get("Proxy-Authorization"), "Basic ")

	data, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return ""
	}

	return string(data)
}
