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

// Package ipauth provides a martian.Modifier that sets auth based on IP.
package ipauth

import (
	"net"
	"net/http"

	"github.com/google/martian"
)

// Modifier is the IP authentication modifier.
type Modifier struct {
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier
}

// NewModifier returns a new IP authentication modifier.
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
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		ip = req.RemoteAddr
	}

	ctx.Auth.ID = ip

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

// ModifyResponse runs cresmod.ModifyResponse and checks ctx.Auth.Error for
// application specific auth failure.
//
// If an error is returned from resmod.ModifyResponse it is returned.
func (m *Modifier) ModifyResponse(ctx *martian.Context, res *http.Response) error {
	if m.resmod != nil {
		if err := m.resmod.ModifyResponse(ctx, res); err != nil {
			return err
		}
	}

	if err := ctx.Auth.Error; err != nil {
		ctx.Auth.Reset()
		ctx.SkipRoundTrip = false
		return err
	}

	return nil
}
