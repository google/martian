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
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/martian/v3/log"
)

type exportHandler struct {
	logger *Logger
}

type resetHandler struct {
	logger *Logger
}

type dumpHandler struct {
	logger   *Logger
	filePath string
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

// NewDumpHandler returns an http.Handler for dumping log entries to file.
func NewDumpHandler(l *Logger, path string) http.Handler {
	return &dumpHandler{
		logger:   l,
		filePath: path,
	}
}

type fileResponseWriter struct {
	file  io.Writer
	multi io.Writer
}

func (w *fileResponseWriter) Write(b []byte) (int, error) {
	return w.multi.Write(b)
}

func newFileResponseWriter(file io.Writer) io.Writer {
	multi := io.MultiWriter(file)
	return &fileResponseWriter{
		file:  file,
		multi: multi,
	}
}

func (h *dumpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		rw.Header().Add("Allow", "GET")
		rw.WriteHeader(http.StatusMethodNotAllowed)
		log.Errorf("har.ServeHTTP: method not allowed: %s", req.Method)
		return
	}
	log.Debugf("exportHandler.ServeHTTP: writing HAR logs to ResponseWriter")
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	dir := filepath.Dir(h.filePath)
	err := os.MkdirAll(dir, 0764)
	if err != nil {
		log.Errorf("mkdir", err)
	}

	file, err := os.Create(h.filePath)
	defer file.Close()
	if err != nil {
		log.Errorf("create file", err)
		return
	}

	writer := newFileResponseWriter(file)
	hl := h.logger.Export()
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(hl)
}

// ServeHTTP writes the log in HAR format to the response body.
func (h *exportHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		rw.Header().Add("Allow", "GET")
		rw.WriteHeader(http.StatusMethodNotAllowed)
		log.Errorf("har.ServeHTTP: method not allowed: %s", req.Method)
		return
	}
	log.Debugf("exportHandler.ServeHTTP: writing HAR logs to ResponseWriter")
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	hl := h.logger.Export()
	json.NewEncoder(rw).Encode(hl)
}

// ServeHTTP resets the log, which clears its entries.
func (h *resetHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !(req.Method == "POST" || req.Method == "DELETE") {
		rw.Header().Add("Allow", "POST")
		rw.Header().Add("Allow", "DELETE")
		rw.WriteHeader(http.StatusMethodNotAllowed)
		log.Errorf("har: method not allowed: %s", req.Method)
		return
	}

	v, err := parseBoolQueryParam(req.URL.Query(), "return")
	if err != nil {
		log.Errorf("har: invalid value for return param: %s", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if v {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		hl := h.logger.ExportAndReset()
		json.NewEncoder(rw).Encode(hl)
	} else {
		h.logger.Reset()
		rw.WriteHeader(http.StatusNoContent)
	}

	log.Infof("resetHandler.ServeHTTP: HAR logs cleared")
}

func parseBoolQueryParam(params url.Values, name string) (bool, error) {
	if params[name] == nil {
		return false, nil
	}
	v, err := strconv.ParseBool(params.Get("return"))
	if err != nil {
		return false, err
	}
	return v, nil
}
