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

package auth

import "github.com/google/martian/session"

const key = "auth.Context"

// Context contains authentication information.
type Context struct {
	ctx *session.Context
	err error
}

// FromContext retrieves the auth.Context from the session.
func FromContext(ctx *session.Context) *Context {
	v, ok := ctx.Get(key)
	if !ok {
		actx := &Context{
			ctx: ctx,
		}
		ctx.Set(key, actx)
		return actx
	}

	return v.(*Context)
}

// ID returns the session ID.
func (ctx *Context) ID() string {
	return ctx.ctx.SessionID()
}

// SetID sets the session ID.
func (ctx *Context) SetID(id string) {
	ctx.err = nil

	if id == "" {
		return
	}

	ctx.ctx.SetSessionID(id)
}

// SetError sets the authentication error and resets the session ID.
func (ctx *Context) SetError(err error) {
	ctx.ctx.SetSessionID("")
	ctx.err = err
}

// Error returns the authentication error.
func (ctx *Context) Error() error {
	return ctx.err
}
