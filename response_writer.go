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
	ow          io.WriteCloser // original writer
	cw          io.WriteCloser // current writer
	hdr         http.Header
	chunked     bool
	wroteHeader bool
	closing     bool
}

type writeCloser interface {
	io.Writer
	io.Closer
}

type nopWriteCloser struct {
	io.Writer
}

func (wc *nopWriteCloser) Close() error {
	return nil
}

// newResponseWriter returns a new http.ResponseWriter.
func newResponseWriter(w io.Writer, closing bool) *responseWriter {
	wc, ok := w.(writeCloser)
	if !ok {
		wc = &nopWriteCloser{w}
	}

	return &responseWriter{
		ow:      wc,
		cw:      wc,
		hdr:     http.Header{},
		closing: closing,
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

	return rw.cw.Write(b)
}

// WriteHeader writes the status line and headers.
func (rw *responseWriter) WriteHeader(status int) {
	if rw.wroteHeader {
		return
	}
	rw.wroteHeader = true

	fmt.Fprintf(rw.ow, "HTTP/1.1 %d %s\r\n", status, http.StatusText(status))

	if rw.closing {
		rw.hdr.Set("Connection", "close")
	}

	if rw.hdr.Get("Content-Length") == "" {
		rw.hdr.Set("Transfer-Encoding", "chunked")
		rw.chunked = true
		rw.cw = httputil.NewChunkedWriter(rw.ow)
	}

	rw.hdr.Write(rw.ow)
	rw.ow.Write([]byte("\r\n"))
}

// Close closes the underlying writer if it is also an io.Closer.
func (rw *responseWriter) Close() error {
	if err := rw.cw.Close(); err != nil {
		return err
	}

	if rw.chunked {
		rw.ow.Write([]byte("\r\n"))
	}

	return nil
}
