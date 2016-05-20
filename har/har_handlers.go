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

package har

import (
	"encoding/json"
	"net/http"
)

type exportHandler struct {
	logger *Logger
}

type resetHandler struct {
	logger *Logger
}

// NewExportHandler returns an http.Handler for requesting HAR logs.
func NewExportHandler(l *Logger) http.Handler {
	return &exportHandler{
		logger: l,
	}
}

// NewResetHandler returns an http.Handler for clearing in-memory log entries.
func NewResetHandler(l *Logger) http.Handler {
	return &resetHandler{
		logger: l,
	}
}

// ServeHTTP writes the log in HAR format to the response body.
func (h *exportHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	hl := h.logger.Export()
	json.NewEncoder(rw).Encode(hl)
}

// ServeHTTP resets the log, which clears its entries.
func (h *resetHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.logger.Reset()

	rw.WriteHeader(http.StatusNoContent)
}
