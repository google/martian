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

package port

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

type Modifier struct {
	port int
}

type modifierJSON struct {
	Port  int                  `json:"port"`
	Scope []parse.ModifierType `json:"scope"`
}

func NewModifier(port int) martian.RequestModifier {
	return &Modifier{
		port: port,
	}
}

func (m *Modifier) ModifyRequest(req *http.Request) error {
	host := req.URL.Host
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err != nil {
			return err
		}
		host = h
	}

	hp := net.JoinHostPort(host, strconv.Itoa(m.port))
	req.URL.Host = hp
	req.Header.Set("Host", hp)

	return nil
}

func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewModifier(msg.Port)

	return parse.NewResult(mod, msg.Scope)
}
