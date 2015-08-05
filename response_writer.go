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

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// responseWriter is a lightweight http.ResponseWriter designed to allow
// Martian to support in-proxy endpoints.
//
// responseWriter does not support all of the functionality of the net/http
// ResponseWriter; in particular, it does not sniff Content-Types.
type responseWriter struct {
	ow          io.Writer // original writer
	w           io.Writer // current writer
	hdr         http.Header
	wroteHeader bool
}

// newResponseWriter returns a new http.ResponseWriter.
func newResponseWriter(w io.Writer) *responseWriter {
	return &responseWriter{
		ow:  w,
		w:   w,
		hdr: http.Header{},
	}
}

// Header returns the headers for the response writer.
func (rw *responseWriter) Header() http.Header {
	return rw.hdr
}

// Write writes b to the response body; if the header has yet to be written it
// will write that before the body.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}

	return rw.w.Write(b)
}

// WriteHeader writes the status line and headers.
func (rw *responseWriter) WriteHeader(status int) {
	if rw.wroteHeader {
		return
	}
	rw.wroteHeader = true

	fmt.Fprintf(rw.w, "HTTP/1.1 %d %s\r\n", status, http.StatusText(status))

	var chunked bool
	if rw.hdr.Get("Content-Length") == "" {
		rw.hdr.Set("Transfer-Encoding", "chunked")
		chunked = true
	}
	rw.hdr.Write(rw.w)

	rw.w.Write([]byte("\r\n"))

	if chunked {
		rw.w = httputil.NewChunkedWriter(rw.w)
	}
}

// Close closes the underlying writer if it is also an io.Closer.
func (rw *responseWriter) Close() error {
	var err error

	wc, ok := rw.w.(io.Closer)
	if ok {
		err = wc.Close()
	}

	rw.ow.Write([]byte("\r\n"))

	return err
}
