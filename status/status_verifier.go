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

package status

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

// Verifier verifies the status codes of all responses.
type Verifier struct {
	statusCode int
}

type verifierJSON struct {
	StatusCode int                  `json:"statusCode"`
	Scope      []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("status.Verifier", verifierFromJSON)
}

// NewVerifier returns a new status.Verifier for statusCode.
func NewVerifier(statusCode int) *Verifier {
	return &Verifier{
		statusCode: statusCode,
	}
}

// ModifyResponse verifies that the status code for all requests
// matches statusCode.
func (v *Verifier) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)

	if res.StatusCode != v.statusCode {
		ev := verify.ResponseError("status.Verifier", res)

		ev.Actual = strconv.Itoa(res.StatusCode)
		ev.Expected = strconv.Itoa(v.statusCode)

		return verify.ForContext(ctx, ev)
	}

	return nil
}

// verifierFromJSON builds a status.Verifier from JSON.
//
// Example JSON:
// {
//   "status.Verifier": {
//     "scope": ["response"],
//     "statusCode": 401
//   }
// }
func verifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &verifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewVerifier(msg.StatusCode), msg.Scope)
}
