// Copyright 2016 Google Inc. All rights reserved.
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

// Package servemux contains a filter that executes modifiers when there is a
// pattern match in a mux.
package servemux

import (
	"net/http"

	"github.com/google/martian"
)

var noop = martian.Noop("mux.Filter")

// Filter is a modifier that executes mod if a pattern is matched in mux.
type Filter struct {
	mux    *http.ServeMux
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier
}

// SetRequestModifier sets the request modifier of filter.
func (f *Filter) SetRequestModifier(reqmod martian.RequestModifier) {
	if reqmod == nil {
		f.reqmod = noop
		return
	}

	f.reqmod = reqmod
}

// SetResponseModifier sets the response modifier of filter.
func (f *Filter) SetResponseModifier(resmod martian.ResponseModifier) {
	if resmod == nil {
		f.resmod = noop
		return
	}

	f.resmod = resmod
}

// NewFilter returns a modifer which runs mod if there is a pattern match in
// mux. The filter will default to http.DefaultServeMux if a mux is not provided.
func NewFilter(mux *http.ServeMux) *Filter {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	return &Filter{
		reqmod: noop,
		resmod: noop,
		mux:    mux,
	}
}

// ModifyRequest executes reqmod iff there is a pattern match in mux.
func (f *Filter) ModifyRequest(req *http.Request) error {
	if _, pattern := f.mux.Handler(req); pattern != "" {
		return f.reqmod.ModifyRequest(req)
	}

	return nil
}

// ModifyResponse executes resmod iff there is pattern match with res.Request in mux.
func (f *Filter) ModifyResponse(res *http.Response) error {
	if _, pattern := f.mux.Handler(res.Request); pattern != "" {
		return f.resmod.ModifyResponse(res)
	}

	return nil
}
