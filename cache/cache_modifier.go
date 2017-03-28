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

package cache_manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

func init() {
	parse.Register("cache.Modifier", modifierFromJSON)
}

type modifierJSON struct {
	Mode string `json:"mode"`
}

type cacheModifier struct {
	cache_database *cacheDatabase
}

type replayModifier struct {
	cache_database *cacheDatabase
}

type HeaderField struct {
	Key    string
	Values []string
}

type SerializedHttpResponse struct {
	ResponseCode int
	Body         string
	Headers      []HeaderField
}

// go binary encoder
func EncodeResponse(m *http.Response) string {
	r := SerializedHttpResponse{}
	log.Printf("Response in EncodeResponse: %v", m)

	bb, err := ioutil.ReadAll(m.Body)
	if err != nil {
		log.Fatal(err)
	}
	m.Body = ioutil.NopCloser(bytes.NewReader(bb))

	r.Body = string(bb)
	r.ResponseCode = m.StatusCode
	for key, headers := range m.Header {
		r.Headers = append(r.Headers, HeaderField{
			Key:    key,
			Values: headers})
	}

	b, err := json.Marshal(r)
	if err != nil {
		log.Fatalf("Error marshalling %v", err)
	}
	return string(b)
}

// go binary decoder
func DecodeResponse(str string, res *http.Response) error {
	r := SerializedHttpResponse{}
	err := json.Unmarshal([]byte(str), &r)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Unmartial result: %v", r)

	res.Body = ioutil.NopCloser(strings.NewReader(r.Body))
	res.Header = make(map[string][]string)
	for _, hi := range r.Headers {
		res.Header[hi.Key] = hi.Values
	}
	res.StatusCode = r.ResponseCode

	log.Printf("Response in FromGOB64: %v", res)

	return nil
}

func (m *replayModifier) ModifyRequest(req *http.Request) error {
	_, ok := m.cache_database.DBMap[req.RequestURI]
	if ok {
		ctx := martian.NewContext(req)
		ctx.SkipRoundTrip()
		return nil
	} else {
		return nil
	}
}

// It loads the the key/value map
func (m *replayModifier) ModifyResponse(res *http.Response) error {
	log.Printf("Replay ModifyResponse handling requestURI: %v", res.Request.RequestURI)
	s, ok := m.cache_database.DBMap[res.Request.RequestURI]
	if !ok {
		log.Printf("Unable to retrieve response for: %v", res.Request.RequestURI)
		return nil
	}
	if err := DecodeResponse(s, res); err != nil {
		return err
	}
	return nil
}

func (m *cacheModifier) ModifyRequest(req *http.Request) error {
	return nil
}

// It stores the the key/value map
func (m *cacheModifier) ModifyResponse(res *http.Response) error {
	enc := EncodeResponse(res)
	log.Print("Encoded into:", enc)
	m.cache_database.DBMap[res.Request.RequestURI] = enc

	return nil
}

// NewModifier returns a modifier that will set the header at name with
// the given value for both requests and responses. If the header name already
// exists all values will be overwritten.
func NewCacheModifier() martian.RequestResponseModifier {
	return &cacheModifier{
		cache_database: GetTheCacheDatabase(),
	}
}

func NewReplayModifier() martian.RequestResponseModifier {
	return &replayModifier{
		cache_database: GetTheCacheDatabase(),
	}
}

// modifierFromJSON takes a JSON message as a byte slice and returns
// a cacheModifer and an error.
//
// Example JSON configuration message:
// {"cache.Modifier": { "mode": "cache"}}
// {"cache.Modifier": { "mode": "replay"}}
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	if msg.Mode == "cache" {
		sp := []parse.ModifierType{"response"}
		modifier := NewCacheModifier()
		return parse.NewResult(modifier, sp)
	} else if msg.Mode == "replay" {
		sp := []parse.ModifierType{"request", "response"}
		modifier := NewReplayModifier()
		return parse.NewResult(modifier, sp)
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported mode %v", msg.Mode))
	}
}
