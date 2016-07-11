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
	"encoding/json"
	"net/http"

	"github.com/google/martian"
	"github.com/google/martian/parse"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/verify"
)

var noop = martian.Noop("header.Filter")

// Filter filters requests and responses based on header name and value.
type Filter struct {
	name, value string
	reqmod      martian.RequestModifier
	resmod      martian.ResponseModifier
	elsereqmod  martian.RequestModifier
	elseresmod  martian.ResponseModifier
}

type filterJSON struct {
	Name         string               `json:"name"`
	Value        string               `json:"value"`
	Modifier     json.RawMessage      `json:"modifier"`
	ElseModifier json.RawMessage      `json:"else"`
	Scope        []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("header.Filter", filterFromJSON)
}

// NewFilter builds a new header filter.
func NewFilter(name, value string) *Filter {
	return &Filter{
		name:       http.CanonicalHeaderKey(name),
		value:      value,
		reqmod:     noop,
		resmod:     noop,
		elsereqmod: noop,
		elseresmod: noop,
	}
}

// SetRequestModifier sets the request modifier of filter.
func (f *Filter) SetRequestModifier(reqmod martian.RequestModifier, elsemod martian.RequestModifier) {
	if reqmod == nil {
		f.reqmod = noop
		return
	}

	f.reqmod = reqmod

	if elsemod == nil {
		f.elsereqmod = noop
		return
	}

	f.elsereqmod = elsemod
}

// SetResponseModifier sets the response modifier of filter.
func (f *Filter) SetResponseModifier(resmod martian.ResponseModifier, elsemod martian.ResponseModifier) {
	if resmod == nil {
		f.resmod = noop
		return
	}

	f.resmod = resmod

	if elsemod == nil {
		f.elseresmod = noop
		return
	}

	f.elseresmod = elsemod
}

// ModifyRequest runs reqmod iff req has a header with name matching value.
func (f *Filter) ModifyRequest(req *http.Request) error {
	h := proxyutil.RequestHeader(req)

	vs, ok := h.All(f.name)
	if !ok {
		return nil
	}

	for _, v := range vs {
		if v == f.value {
			return f.reqmod.ModifyRequest(req)
		}
	}

	if f.elsereqmod != nil {
		return f.elsereqmod.ModifyRequest(req)
	}

	return nil
}

// ModifyResponse runs resmod iff res has a header with name matching value.
func (f *Filter) ModifyResponse(res *http.Response) error {
	h := proxyutil.ResponseHeader(res)

	vs, ok := h.All(f.name)
	if !ok {
		return nil
	}

	for _, v := range vs {
		if v == f.value {
			return f.resmod.ModifyResponse(res)
		}
	}

	if f.elseresmod != nil {
		return f.elseresmod.ModifyResponse(res)
	}

	return nil
}

// VerifyRequests returns an error containing all the verification errors
// returned by request verifiers.
func (f *Filter) VerifyRequests() error {
	reqv, ok := f.reqmod.(verify.RequestVerifier)
	if !ok {
		return nil
	}

	return reqv.VerifyRequests()
}

// VerifyResponses returns an error containing all the verification errors
// returned by response verifiers.
func (f *Filter) VerifyResponses() error {
	resv, ok := f.resmod.(verify.ResponseVerifier)
	if !ok {
		return nil
	}

	return resv.VerifyResponses()
}

// ResetRequestVerifications resets the state of the contained request verifiers.
func (f *Filter) ResetRequestVerifications() {
	if reqv, ok := f.reqmod.(verify.RequestVerifier); ok {
		reqv.ResetRequestVerifications()
	}
}

// ResetResponseVerifications resets the state of the contained request verifiers.
func (f *Filter) ResetResponseVerifications() {
	if resv, ok := f.resmod.(verify.ResponseVerifier); ok {
		resv.ResetResponseVerifications()
	}
}

// filterFromJSON builds a header.Filter from JSON.
//
// Example JSON:
// {
//   "scope": ["request", "result"],
//   "name": "Martian-Testing",
//   "value": "true",
//   "modifier": { ... }
//   "else": {...}
// }
func filterFromJSON(b []byte) (*parse.Result, error) {
	msg := &filterJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	filter := NewFilter(msg.Name, msg.Value)

	m, err := parse.FromJSON(msg.Modifier)
	if err != nil {
		return nil, err
	}

	reqmod := m.RequestModifier()
	resmod := m.ResponseModifier()

	em, err := parse.FromJSON(msg.ElseModifier)
	if err != nil {
		return nil, err
	}

	elsereqmod := em.RequestModifier()
	elseresmod := em.ResponseModifier()

	filter.SetRequestModifier(reqmod, elsereqmod)
	filter.SetResponseModifier(resmod, elseresmod)

	return parse.NewResult(filter, msg.Scope)
}
