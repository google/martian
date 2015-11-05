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

package trafficshape

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/martian/log"
)

// Handler configures a trafficshape.Listener.
type Handler struct {
	l *Listener
}

// NewHandler returns an http.Handler to configure traffic shaping.
func NewHandler(l *Listener) *Handler {
	return &Handler{
		l: l,
	}
}

// ServeHTTP configures latency and bandwidth constraints.
//
// The "latency" query string parameter accepts a duration string in any format
// supported by time.ParseDuration.
// The "up" and "down" query string parameters accept integers as bits per
// second to be used for read and write throughput.
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("trafficshape: configuration request")

	latency := req.FormValue("latency")
	if latency != "" {
		d, err := time.ParseDuration(latency)
		if err != nil {
			log.Errorf("trafficshape: invalid latency parameter: %v", err)
			http.Error(rw, fmt.Sprintf("invalid duration: %s", latency), 400)
			return
		}

		h.l.SetLatency(d)
	}

	up := req.FormValue("up")
	if up != "" {
		br, err := strconv.ParseInt(up, 10, 64)
		if err != nil {
			log.Errorf("trafficshape: invalid up parameter: %v", err)
			http.Error(rw, fmt.Sprintf("invalid upstream: %s", up), 400)
			return
		}

		h.l.SetWriteBitrate(br)
	}

	down := req.FormValue("down")
	if down != "" {
		br, err := strconv.ParseInt(down, 10, 64)
		if err != nil {
			log.Errorf("trafficshape: invalid down parameter: %v", err)
			http.Error(rw, fmt.Sprintf("invalid downstream: %s", down), 400)
			return
		}

		h.l.SetReadBitrate(br)
	}

	log.Debugf("trafficshape: configured successfully")
}
