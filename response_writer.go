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
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
)

var errHijacked = errors.New("martian: connection has already been hijacked")

// responseWriter is a lightweight http.ResponseWriter designed to allow
// Martian to support in-proxy endpoints.
//
// responseWriter does not support all of the functionality of the net/http
// ResponseWriter; in particular, it does not sniff Content-Types.
type responseWriter struct {
	conn        net.Conn
	bw          *bufio.ReadWriter
	hdr         http.Header
	hijacked    bool
	chunked     bool
	wroteHeader bool
	closing     bool
}

// newResponseWriter returns a new http.ResponseWriter.
func newResponseWriter(conn net.Conn, bw *bufio.ReadWriter, closing bool) *responseWriter {
	return &responseWriter{
		conn:    conn,
		bw:      bw,
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
	if rw.hijacked {
		return 0, errHijacked
	}

	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}

	return rw.bw.Write(b)
}

// WriteHeader writes the status line and headers.
func (rw *responseWriter) WriteHeader(status int) {
	if rw.wroteHeader || rw.hijacked {
		return
	}
	rw.wroteHeader = true

	fmt.Fprintf(rw.bw, "HTTP/1.1 %d %s\r\n", status, http.StatusText(status))

	if rw.closing {
		rw.hdr.Set("Connection", "close")
	}

	if rw.hdr.Get("Content-Length") == "" {
		rw.hdr.Set("Transfer-Encoding", "chunked")
		rw.chunked = true

		rw.hdr.Write(rw.bw)
		rw.bw.Write([]byte("\r\n"))
		rw.bw.Flush()

		rw.bw.Writer.Reset(httputil.NewChunkedWriter(rw.conn))
		return
	}

	rw.hdr.Write(rw.bw)
	rw.bw.Write([]byte("\r\n"))
}

// Close writes the trailing newline for chunked responses.
func (rw *responseWriter) Close() error {
	if rw.hijacked {
		return errHijacked
	}

	if rw.chunked {
		defer rw.bw.Writer.Reset(rw.conn)
		rw.bw.Flush()
		rw.conn.Write([]byte("0\r\n\r\n"))
	}

	return nil
}

// Hijack disconnects the underlying connection from the ResponseWriter and
// returns it to the handler.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.hijacked {
		return nil, nil, errHijacked
	}
	rw.hijacked = true

	return rw.conn, rw.bw, nil
}
