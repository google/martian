// Copyright 2021 Google Inc. All rights reserved.
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

package h2

import (
	"fmt"
	"net/url"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

// Direction indicates the direction of the traffic flow.
type Direction uint8

const (
	// ClientToServer indicates traffic flowing from client-to-server.
	ClientToServer Direction = iota
	// ServerToClient indicates traffic flowing from server-to-client.
	ServerToClient
)

// StreamProcessorFactory is implemented by clients that wish to observe or edit HTTP/2 frames
// flowing through the proxy. It creates a pair of processors for the bidirectional stream. A
// processor consumes frames then calls the corresponding sink methods to forward frames to the
// destination, modifying the frame if needed.
//
// Returns the client-to-server and server-to-client processors. Nil values are safe to return and
// no processing occurs in such cases. NOTE: an interface may have a non-nil type with a nil value.
// Such values are treated as valid processors.
//
// Concurrency: there is a separate client-to-server and server-to-client thread. Calls against
// the `ClientToServer` sink must be made on the client-to-server thread and calls against
// the `ServerToClient` sink must be made on the server-to-client thread. Implementors should
// guard interactions across threads.
type StreamProcessorFactory func(url *url.URL, sinks *Processors) (Processor, Processor)

// Processors encapsulates the two traffic receiving endpoints.
type Processors struct {
	cToS, sToC Processor
}

// ForDirection returns the processor receiving traffic in the given direction.
func (s *Processors) ForDirection(dir Direction) Processor {
	switch dir {
	case ClientToServer:
		return s.cToS
	case ServerToClient:
		return s.sToC
	}
	panic(fmt.Sprintf("invalid direction: %v", dir))
}

// Processor accepts the possible stream frames.
//
// This API abstracts away some of the lower level HTTP/2 mechanisms.
// CONTINUATION frames are appropriately buffered and turned into Header calls and Header or
// PushPromise calls are split into CONTINUATION frames when needed.
//
// The proxy handles WINDOW_UPDATE frames and flow control, managing it independently for both
// endpoints.
type Processor interface {
	DataFrameProcessor
	HeaderProcessor
	PriorityFrameProcessor
	RSTStreamProcessor
	PushPromiseProcessor
}

// DataFrameProcessor processes data frames.
type DataFrameProcessor interface {
	Data(data []byte, streamEnded bool) error
}

// HeaderProcessor processes headers, abstracting out continuations.
type HeaderProcessor interface {
	Header(
		headers []hpack.HeaderField,
		streamEnded bool,
		priority http2.PriorityParam,
	) error
}

// PriorityFrameProcessor processes priority frames.
type PriorityFrameProcessor interface {
	Priority(http2.PriorityParam) error
}

// RSTStreamProcessor processes RSTStream frames.
type RSTStreamProcessor interface {
	RSTStream(http2.ErrCode) error
}

// PushPromiseProcessor processes push promises, abstracting out continuations.
type PushPromiseProcessor interface {
	PushPromise(promiseID uint32, headers []hpack.HeaderField) error
}

// relayAdapter implements the Processor interface by delegating to an underlying relay.
type relayAdapter struct {
	id    uint32
	relay *relay
}

func (r *relayAdapter) Data(data []byte, streamEnded bool) error {
	return r.relay.data(r.id, data, streamEnded)
}

func (r *relayAdapter) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	return r.relay.header(r.id, headers, streamEnded, priority)
}

func (r *relayAdapter) Priority(priority http2.PriorityParam) error {
	r.relay.priority(r.id, priority)
	return nil
}

func (r *relayAdapter) RSTStream(errCode http2.ErrCode) error {
	r.relay.rstStream(r.id, errCode)
	return nil
}

func (r *relayAdapter) PushPromise(promiseID uint32, headers []hpack.HeaderField) error {
	return r.relay.pushPromise(r.id, promiseID, headers)
}
