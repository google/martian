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
	"net/http"

	"github.com/google/martian"
)

// NewViaModifier sets the Via header.
//
// If Via is already present, via is appended to the existing value.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-9.9
func NewViaModifier(via string) martian.RequestModifier {
	return martian.RequestModifierFunc(
		func(req *http.Request) error {
			if v := req.Header.Get("Via"); v != "" {
				via = v + ", " + via
			}

			req.Header.Set("Via", via)

			return nil
		})
}
