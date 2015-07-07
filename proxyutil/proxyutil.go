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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

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
		StatusCode: code,
		Status:     fmt.Sprintf("%d %s", code, http.StatusText(code)),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       rc,
		Request:    req,
	}
}

// NewErrorResponse builds new HTTP error responses.
func NewErrorResponse(code int, err error, req *http.Request) *http.Response {
	res := NewResponse(code, strings.NewReader(err.Error()), req)
	res.Header.Set("Content-Type", "text/plain; charset=utf-8")
	res.ContentLength = int64(len(err.Error()))

	return res
}
