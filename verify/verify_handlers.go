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

package verify

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian/log"
)

// Handler is an http.Handler that returns the request and response errors of
// the verification.
type Handler struct {
	v *Verification
}

// ResetHandler is an http.Handler that resets the request and response errors
// of the verification.
type ResetHandler struct {
	v *Verification
}

type errorsJSON struct {
	Errors []*ErrorValue `json:"errors"`
}

// NewHandler returns an http.Handler for requesting verification errors.
func NewHandler(v *Verification) *Handler {
	return &Handler{
		v: v,
	}
}

// NewResetHandler returns an http.Handler for resetting verification errors.
func NewResetHandler(v *Verification) *ResetHandler {
	return &ResetHandler{
		v: v,
	}
}

// ServeHTTP writes a JSON response containing a list of verification errors
// that have occurred.
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	if req.Method != "GET" {
		rw.Header().Set("Allow", "GET")
		rw.WriteHeader(405)
		log.Errorf("verify: invalid request method: %s", req.Method)
		return
	}

	ej := &errorsJSON{
		Errors: make([]*ErrorValue, 0, len(h.errs)),
	}

	for _, err := range h.v.Errors() {
		ej.Errors = append(ej.Errors, err.Get())
	}

	json.NewEncoder(rw).Encode(ej)
}

// ServeHTTP resets the verification errors.
func (h *ResetHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.Header().Set("Allow", "POST")
		rw.WriteHeader(405)
		log.Errorf("verify: invalid request method: %s", req.Method)
		return
	}

	h.v.Reset()

	rw.WriteHeader(204)
}
