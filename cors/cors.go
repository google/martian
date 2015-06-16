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

// Package cors provides CORS support for http.Handlers.
package cors

import (
	"net/http"
)

type handler struct {
	handler http.Handler
}

// NewHandler wraps an existing http.Handler allowing it to be requested via CORS.
func NewHandler(h http.Handler) http.Handler {
	return &handler{
		handler: h,
	}
}

// ServeHTTP determines if a request is a CORS request (normal or preflight)
// and sets the appropriate Access-Control-Allow-* headers. It will send the
// request to the underlying handler in all cases, except for a preflight
// (OPTIONS) request.
func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Definitely not a CORS request, send it directly to handler.
	if req.Header.Get("Origin") == "" {
		h.handler.ServeHTTP(rw, req)
		return
	}

	rw.Header().Set("Access-Control-Allow-Origin", "*")

	acrm := req.Header.Get("Access-Control-Request-Method")
	rw.Header().Set("Access-Control-Allow-Methods", acrm)

	if acrh := req.Header.Get("Access-Control-Request-Headers"); acrh != "" {
		rw.Header().Set("Access-Control-Allow-Headers", acrh)
	}

	// Preflight request, don't bother sending it to the handler.
	if req.Method == "OPTIONS" {
		return
	}

	h.handler.ServeHTTP(rw, req)
}
