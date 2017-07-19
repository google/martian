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

// Package body allows for the replacement of message body on responses.
package body

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/martian/parse"
)

func init() {
	parse.Register("body.Modifier", modifierFromJSON)
}

// Modifier substitutes the body on an HTTP response.
type Modifier struct {
	contentType string
	body        []byte
}

type modifierJSON struct {
	ContentType string               `json:"contentType"`
	Body        []byte               `json:"body"` // Body is expected to be a Base64 encoded string.
	Scope       []parse.ModifierType `json:"scope"`
}

// NewModifier constructs and returns a body.Modifier.
func NewModifier(b []byte, contentType string) *Modifier {
	return &Modifier{
		contentType: contentType,
		body:        b,
	}
}

// modifierFromJSON takes a JSON message as a byte slice and returns a
// body.Modifier and an error.
//
// Example JSON Configuration message:
// {
//   "scope": ["request", "response"],
//   "contentType": "text/plain",
//   "body": "c29tZSBkYXRhIHdpdGggACBhbmQg77u/" // Base64 encoded body
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewModifier(msg.Body, msg.ContentType)
	return parse.NewResult(mod, msg.Scope)
}

// ModifyRequest sets the Content-Type header and overrides the request body.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	req.Body.Close()

	req.Header.Set("Content-Type", m.contentType)

	// Reset the Content-Encoding since we know that the new body isn't encoded.
	req.Header.Del("Content-Encoding")

	req.ContentLength = int64(len(m.body))
	req.Body = ioutil.NopCloser(bytes.NewReader(m.body))

	return nil
}

// ModifyResponse sets the Content-Type header and overrides the response body.
func (m *Modifier) ModifyResponse(res *http.Response) error {
	// Replace the existing body, close it first.
	res.Body.Close()

	res.Header.Set("Content-Type", m.contentType)

	// Check for Range request.
	rhv := res.Request.Header.Get("Range")
	if rhv != "" && strings.HasPrefix(rhv, "bytes=") {
		rh := res.Request.Header.Get("Range")
		rh = strings.ToLower(rh)
		rs := strings.Split(strings.TrimLeft(rh, "bytes="), "-")
		if len(rs) != 2 {
			return fmt.Errorf("body.Modifier: invalid Range request header: %s", rh)
		}
		start, err := strconv.Atoi(rs[0])
		if err != nil {
			return err
		}

		end, err := strconv.Atoi(rs[1])
		if err != nil {
			return err
		}

		if start > end {
			return fmt.Errorf("body.Modifier: invalid range request. %d should be greater than %d", end, start)
		}

		br := bufio.NewReader(bytes.NewReader(m.body))
		if _, err := br.Discard(start); err != nil {
			return err
		}

		var segment []byte
		seglen := end - start + 1
		for i := 0; i < seglen; i++ {
			b, err := br.ReadByte()
			if err != nil {
				return err
			}
			segment = append(segment, b)
		}

		res.ContentLength = int64(len(segment))
		res.Body = ioutil.NopCloser(bytes.NewReader(segment))
		res.StatusCode = http.StatusPartialContent
		res.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(m.body)))

		return nil
	}

	// Reset the Content-Encoding since we know that the new body isn't encoded.
	res.Header.Del("Content-Encoding")

	res.ContentLength = int64(len(m.body))
	res.Body = ioutil.NopCloser(bytes.NewReader(m.body))

	return nil
}
