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

// Package static provides a modifier that allow Martian to reurn static files
// local to Martian.
package static

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

type staticModifier struct {
	rootPath string
}

type staticJSON struct {
	RootPath string               `json:"rootPath"`
	Scope    []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("static.Modifier", modifierFromJSON)
}

// NewModifier constructs a staticModifier that takes a path from which to
// serve files from.
func NewModifier(rootPath string) martian.RequestResponseModifier {
	return &staticModifier{
		rootPath: path.Clean(rootPath),
	}
}

// ModifyRequest marks the context to skip the roundtrip.
func (s *staticModifier) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	ctx.SkipRoundTrip()
	return nil
}

// ModifyResponse reads the file rooted at rootPath joined with the request URL
// path.  In the case that the file cannot be found, the response will be a 404.
func (s *staticModifier) ModifyResponse(res *http.Response) error {
	p := filepath.Join(s.rootPath, filepath.Clean(res.Request.URL.Path))
	f, err := os.Open(p)
	switch {
	case os.IsNotExist(err):
		res.StatusCode = http.StatusNotFound
		return err
	case os.IsPermission(err):
		// This is returning a StatusUnauthorized to reflect that the Martian does
		// not have the appropriate permissions on the local file system.  This is a
		// deviation from the standard assumption around an HTTP 401 response.
		res.StatusCode = http.StatusUnauthorized
		return err
	case err != nil:
		res.StatusCode = http.StatusInternalServerError
		return err
	}

	res.Body.Close()
	res.Body = f

	res.Header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(p)))

	return nil
}

func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &staticJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewModifier(msg.RootPath), msg.Scope)
}
