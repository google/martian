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

package header

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/martian"
)

const viaLoopKey = "via.LoopDetection"

type viaModifier struct {
	requestedBy string
}

// NewViaModifier returns a new Via modifier.
func NewViaModifier(requestedBy string) martian.RequestResponseModifier {
	return &viaModifier{
		requestedBy: requestedBy,
	}
}

// ModifyRequest sets the Via header and provides loop-detection. If Via is
// already present, it will be appended to the existing value. If a loop is
// detected an error is added to the context and the request round trip is
// skipped.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-9.9
func (m *viaModifier) ModifyRequest(req *http.Request) error {
	via := fmt.Sprintf("%d.%d %s", req.ProtoMajor, req.ProtoMinor, m.requestedBy)

	if v := req.Header.Get("Via"); v != "" {
		if strings.Contains(v, m.requestedBy) {
			err := fmt.Errorf("via: detected request loop, header contains %s", m.requestedBy)

			ctx := martian.Context(req)
			ctx.Set(viaLoopKey, err)
			ctx.SkipRoundTrip()

			return err
		}

		via = fmt.Sprintf("%s, %s", v, via)
	}

	req.Header.Set("Via", via)

	return nil
}

// ModifyResponse sets the status code to 400 Bad Request if a loop was
// detected in the request.
func (m *viaModifier) ModifyResponse(res *http.Response) error {
	ctx := martian.Context(res.Request)

	if err, _ := ctx.Get(viaLoopKey); err != nil {
		res.StatusCode = 400
		res.Status = http.StatusText(400)

		return err.(error)
	}

	return nil
}
