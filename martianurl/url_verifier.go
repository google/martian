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

package martianurl

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

func init() {
	parse.Register("url.Verifier", verifierFromJSON)
}

// Verifier verifies the structure of URLs.
type Verifier struct {
	url *url.URL
}

type verifierJSON struct {
	Scheme string               `json:"scheme"`
	Host   string               `json:"host"`
	Path   string               `json:"path"`
	Query  string               `json:"query"`
	Scope  []parse.ModifierType `json:"scope"`
}

// NewVerifier returns a new URL verifier.
func NewVerifier(url *url.URL) *Verifier {
	return &Verifier{
		url: url,
	}
}

// ModifyRequest verifies that the request URL matches all parts of url. If the
// value in url is non-empty it must be an exact match. Each unmatched URL part
// will be treated as a distinct error.
func (v *Verifier) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)

	if v.url.Scheme != "" && v.url.Scheme != req.URL.Scheme {
		ev := verify.RequestError("url.Verifier", req)

		ev.Actual = req.URL.Scheme
		ev.Expected = v.url.Scheme
		ev.MessageFormat = "scheme: got %s, want %s"

		verify.ForContext(ctx, ev)
	}
	if v.url.Host != "" && v.url.Host != req.URL.Host {
		ev := verify.RequestError("url.Verifier", req)

		ev.Actual = req.URL.Host
		ev.Expected = v.url.Host
		ev.MessageFormat = "host: got %s, want %s"

		verify.ForContext(ctx, ev)
	}
	if v.url.Path != "" && v.url.Path != req.URL.Path {
		ev := verify.RequestError("url.Verifier", req)

		ev.Actual = req.URL.Path
		ev.Expected = v.url.Path
		ev.MessageFormat = "path: got %s, want %s"

		verify.ForContext(ctx, ev)
	}
	if v.url.RawQuery != "" && v.url.RawQuery != req.URL.RawQuery {
		ev := verify.RequestError("url.Verifier", req)

		ev.Actual = req.URL.RawQuery
		ev.Expected = v.url.RawQuery
		ev.MessageFormat = "query: got %s, want %s"

		verify.ForContext(ctx, ev)
	}

	return nil
}

// verifierFromJSON builds a martianurl.Verifier from JSON.
//
// Example modifier JSON:
// {
//   "url.Verifier": {
//     "scope": ["request"],
//     "scheme": "https",
//     "host": "www.google.com",
//     "path": "/proxy",
//     "query": "testing=true"
//   }
// }
func verifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &verifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	v := NewVerifier(&url.URL{
		Scheme:   msg.Scheme,
		Host:     msg.Host,
		Path:     msg.Path,
		RawQuery: msg.Query,
	})

	return parse.NewResult(v, msg.Scope)
}
