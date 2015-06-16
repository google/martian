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

// Context contains information for a proxy session.
type Context struct {
	// Auth is the session authentication information.
	Auth	*Auth
	// SkipRoundTrip signals to the proxy that it should not send the request over the wire.
	SkipRoundTrip	bool
}

// Auth contains per session authentication information.
type Auth struct {
	// ID is the identifier for a user.
	ID	string
	// Error is used to signal that ID is required, but is either
	// blank or invalid per the semantics of the modifier.
	Error	error
}

// NewContext returns an empty martian.Context.
func NewContext() *Context {
	return &Context{
		Auth: &Auth{},
	}
}

// Reset resets the auth fields to their zero values.
func (auth *Auth) Reset() {
	auth.ID = ""
	auth.Error = nil
}
