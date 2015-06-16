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

/*
Package proxyutil provides functionality for building proxies.
*/
package proxyutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

// ErrBadFraming is an error returned when it is impossible to determine how to
// properly terminate a request. This often occurs becaue of conflicting
// Content-Length headers or a non-chunked Transfer-Encoding without a
// Content-Length.
var ErrBadFraming = errors.New("bad request framing")

// Hop-by-hop headers as defined by RFC2616.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-7.1.3.1
var HopByHopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

// NewResponse builds new HTTP responses.
// If body is nil, an empty byte.Buffer will be provided to be consistent with
// the guarantees provided by http.Transport and http.Client.
func NewResponse(code int, body io.Reader, req *http.Request) *http.Response {
	if body == nil {
		body = &bytes.Buffer{}
	}

	rc, ok := body.(io.ReadCloser)
	if !ok {
		rc = ioutil.NopCloser(body)
	}

	return &http.Response{
		StatusCode:	code,
		Status:		fmt.Sprintf("%d %s", code, http.StatusText(code)),
		Proto:		"HTTP/1.1",
		ProtoMajor:	1,
		ProtoMinor:	1,
		Header:		http.Header{},
		Body:		rc,
		Request:	req,
	}
}

// NewErrorResponse builds new HTTP error responses.
func NewErrorResponse(code int, err error, req *http.Request) *http.Response {
	res := NewResponse(code, strings.NewReader(err.Error()), req)
	res.Header.Set("Content-Type", "text/plain; charset=utf-8")
	res.ContentLength = int64(len(err.Error()))

	return res
}

// RemoveHopByHopHeaders removes all hop-by-hop headers defined
// by RFC2616 as well as any additional hop-by-hop headers
// specified in the Connection header.
func RemoveHopByHopHeaders(header http.Header) {
	// Additional hop-by-hop headers may be specified in `Connection` headers.
	// http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-9.1
	for _, vs := range header["Connection"] {
		for _, v := range strings.Split(vs, ",") {
			k := http.CanonicalHeaderKey(strings.TrimSpace(v))
			header.Del(k)
		}
	}

	for _, k := range HopByHopHeaders {
		header.Del(k)
	}
}

// SetForwardedHeaders sets the X-Fowarded-Proto and
// X-Forwarded-For headers for the outgoing request.
//
// If X-Forwarded-For is already present, the client IP is appended to the
// existing value.
func SetForwardedHeaders(req *http.Request) {
	req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)

	xff, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		xff = req.RemoteAddr
	}

	if v := req.Header.Get("X-Forwarded-For"); v != "" {
		xff = v + ", " + xff
	}

	req.Header.Set("X-Forwarded-For", xff)
}

// SetViaHeader sets the Via header.
//
// If Via is already present, via is appended to
// the existing value.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-9.9
func SetViaHeader(header http.Header, via string) {
	if v := header.Get("Via"); v != "" {
		via = v + ", " + via
	}

	header.Set("Via", via)
}

// FixBadFraming makes a best effort to fix inconsistencies in the request such
// as multiple Content-Lengths or the lack of Content-Length and improper
// Transfer-Encoding. If it is unable to determine a proper resolution it
// returns ErrBadFraming.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-3.3
func FixBadFraming(header http.Header) error {
	cls := header["Content-Length"]
	if len(cls) > 0 {
		var length string

		// Iterate over all Content-Length headers, splitting any we find with
		// commas, and check that all Content-Lengths are equal.
		for _, ls := range cls {
			for _, l := range strings.Split(ls, ",") {
				// First length, set it as the canonical Content-Length.
				if length == "" {
					length = strings.TrimSpace(l)
					continue
				}

				// Mismatched Content-Lengths.
				if length != strings.TrimSpace(l) {
					return ErrBadFraming
				}
			}
		}

		// All Content-Lengths are equal, remove extras and set it to the canonical
		// value.
		header.Set("Content-Length", length)
	}

	tes := header["Transfer-Encoding"]
	if len(tes) > 0 {
		// Extract the last Transfer-Encoding value, and split on commas.
		last := strings.Split(tes[len(tes)-1], ",")

		// Check that the last, potentially comma-delimited, value is "chunked",
		// else we have no way to determine when the request is finished.
		if strings.TrimSpace(last[len(last)-1]) != "chunked" {
			return ErrBadFraming
		}

		// Transfer-Encoding "chunked" takes precedence over
		// Content-Length.
		header.Del("Content-Length")
	}

	return nil
}
