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

// Package static provides a modifier that allows Martian to return static files
// local to Martian. The static modifier does not support setting explicit path
// mappings via the JSON API.
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

// Modifier is a martian.RequestResponseModifier that routes reqeusts to rootPath
// and serves the assets there, while skipping the HTTP roundtrip.
type Modifier struct {
	rootPath      string
	explicitPaths map[string]string
}

type staticJSON struct {
	RootPath string               `json:"rootPath"`
	Scope    []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("static.Modifier", modifierFromJSON)
}

// NewModifier constructs a static.Modifier that takes a path to serve files from, as well as an optional mapping of request paths to local
// file paths (still rooted at rootPath).
func NewModifier(rootPath string) *Modifier {
	return &Modifier{
		rootPath:      path.Clean(rootPath),
		explicitPaths: make(map[string]string),
	}
}

// ModifyRequest marks the context to skip the roundtrip and downgrades any https requests
// to http.
func (s *Modifier) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	ctx.SkipRoundTrip()

	if req.URL.Scheme == "https" {
		req.URL.Scheme = "http"
		ctx.Session().MarkInsecure()
	}

	return nil
}

// ModifyResponse reads the file rooted at rootPath joined with the request URL
// path. In the case that the the request path is a key in s.explicitPaths, ModifyRequest
// will attempt to open the file located at s.rootPath joined by the value in s.explicitPaths
// (keyed by res.Request.URL.Path).  In the case that the file cannot be found, the response
// will be a 404. ModifyResponse will return a 404 for any path that is defined in s.explictPaths
// and that does not exist locally, even if that file does exist in s.rootPath.
func (s *Modifier) ModifyResponse(res *http.Response) error {
	reqpth := filepath.Clean(res.Request.URL.Path)
	fpth := filepath.Join(s.rootPath, reqpth)

	if _, ok := s.explicitPaths[reqpth]; ok {
		fpth = filepath.Join(s.rootPath, s.explicitPaths[reqpth])
	}

	f, err := os.Open(fpth)
	switch {
	case os.IsNotExist(err):
		res.StatusCode = http.StatusNotFound
		return nil
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

	info, err := f.Stat()
	if err != nil {
		res.StatusCode = http.StatusInternalServerError
		return err
	}
	res.ContentLength = info.Size()
	res.Header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(fpth)))

	return nil
}

// SetExplicitPathMappings sets an optional mapping of request paths to local
// file paths rooted at s.rootPath.
func (s *Modifier) SetExplicitPathMappings(ep map[string]string) {
	s.explicitPaths = ep
}

func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &staticJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewModifier(msg.RootPath), msg.Scope)
}
