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

// Package log provides a Martian modifier that logs the request and response.
package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

// Logger is a modifier that logs requests and responses.
type Logger struct {
	log         func(line string)
	includeBody bool
}

type loggerJSON struct {
	Scope       []parse.ModifierType `json:"scope"`
	IncludeBody bool                 `json:"includeBody"`
}

func init() {
	parse.Register("log.Logger", loggerFromJSON)
}

// NewLogger returns a logger that logs requests and responses, optionally
// logging the body. Log function defaults to martian.Infof.
func NewLogger() *Logger {
	return &Logger{
		log: func(line string) {
			martian.Infof(line)
		},
	}
}

// IncludeBody sets whether to include the request/response body in the log.
func (l *Logger) IncludeBody(includeBody bool) {
	l.includeBody = includeBody
}

// SetLogFunc sets the logging function for the logger.
func (l *Logger) SetLogFunc(logFunc func(line string)) {
	l.log = logFunc
}

// ModifyRequest logs the request, optionally including the body.
//
// The format logged is:
// --------------------------------------------------------------------------------
// Request to http://www.google.com/path?querystring
// --------------------------------------------------------------------------------
// GET /path?querystring HTTP/1.1
// Host: www.google.com
// Connection: close
// Other-Header: values
//
// request content
// --------------------------------------------------------------------------------
func (l *Logger) ModifyRequest(req *http.Request) error {
	b := &bytes.Buffer{}

	fmt.Fprintln(b, "")
	fmt.Fprintln(b, strings.Repeat("-", 80))
	fmt.Fprintf(b, "Request to %s\n", req.URL)
	fmt.Fprintln(b, strings.Repeat("-", 80))

	dump, err := httputil.DumpRequest(req, l.includeBody)
	if err != nil {
		return err
	}
	b.Write(dump)

	fmt.Fprintln(b, "")
	fmt.Fprintln(b, strings.Repeat("-", 80))

	l.log(b.String())

	return nil
}

// ModifyResponse logs the response, optionally including the body.
//
// The format logged is:
// --------------------------------------------------------------------------------
// Response from http://www.google.com/path?querystring
// --------------------------------------------------------------------------------
// HTTP/1.1 200 OK
// Date: Tue, 15 Nov 1994 08:12:31 GMT
// Other-Header: values
//
// response content
// --------------------------------------------------------------------------------
func (l *Logger) ModifyResponse(res *http.Response) error {
	b := &bytes.Buffer{}
	fmt.Fprintln(b, "")
	fmt.Fprintln(b, strings.Repeat("-", 80))
	fmt.Fprintf(b, "Response from %s\n", res.Request.URL)
	fmt.Fprintln(b, strings.Repeat("-", 80))

	dump, err := httputil.DumpResponse(res, l.includeBody)
	if err != nil {
		return err
	}
	b.Write(dump)

	fmt.Fprintln(b, "")
	fmt.Fprintln(b, strings.Repeat("-", 80))

	l.log(b.String())

	return nil
}

// loggerFromJSON builds a logger from JSON.
//
// Example JSON:
// {
//   "log.Logger": {
//     "scope": ["request", "response"],
//		 "includeBody": true
//   }
// }
func loggerFromJSON(b []byte) (*parse.Result, error) {
	msg := &loggerJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	l := NewLogger()
	l.IncludeBody(msg.IncludeBody)

	return parse.NewResult(l, msg.Scope)
}
