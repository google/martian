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

// Package grpc contains gRPC functionality for Martian proxy.
package grpc

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/url"
	"sync/atomic"

	"github.com/golang/snappy"
	"github.com/google/martian/v3/h2"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

// Encoding is the grpc-encoding type. See Content-Coding entry at:
// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests
type Encoding uint8

const (
	// Identity indicates that no compression is used.
	Identity Encoding = iota
	// Gzip indicates that Gzip compression is used.
	Gzip
	// Deflate indicates that Deflate compression is used.
	Deflate
	// Snappy indicates that Snappy compression is used.
	Snappy
)

// ProcessorFactory creates gRPC processors that implement the Processor interface, which abstracts
// away some of the details of the underlying HTTP/2 protocol. A processor must forward
// invocations to the given `server` or `client` processors, which will arrange to have the data
// forwarded to the destination, with possible edits. Nil values are safe to return and no
// processing occurs in such cases. NOTE: an interface may have a non-nil type with a nil value.
// Such values are treated as valid processors.
type ProcessorFactory func(url *url.URL, server, client Processor) (Processor, Processor)

// AsStreamProcessorFactory converts a ProcessorFactory into a StreamProcessorFactory. It creates
// an adapter that abstracts HTTP/2 frames into a representation that is closer to gRPC.
func AsStreamProcessorFactory(f ProcessorFactory) h2.StreamProcessorFactory {
	return func(url *url.URL, sinks *h2.Processors) (h2.Processor, h2.Processor) {
		var cToS, sToC h2.Processor

		// A grpc.Processor is translated into an h2.Processor in layers.
		//
		// adapter → processor → emitter → sink
		//    \_____________________________↗
		//
		// * The adapter wraps the grpc.Processor interface so that it conforms with h2.Processor. It
		//   performs some processing to translate HTTP/2 frames into gRPC concepts. Frames that are
		//   not relevant to gRPC are forwarded directly to the sink.
		// * The processor is the gRPC processing logic provided by the client factory.
		// * The emitter wraps an h2.Processor sink and translates the processed gRPC data into HTTP/2
		//   frames.
		cToSEmitter := &emitter{sink: sinks.ForDirection(h2.ClientToServer)}
		sToCEmitter := &emitter{sink: sinks.ForDirection(h2.ServerToClient)}
		cToSProcessor, sToCProcessor := f(url, cToSEmitter, sToCEmitter)

		// enabled indicates whether the stream should be processed as gRPC. It is shared between the
		// the two adapters because its detection is on a client-to-server HEADER frame and the state
		// applies bidirectionally.
		enabled := int32(0)
		if cToSProcessor != nil {
			cToSEmitter.adapter = &adapter{
				enabled:   &enabled,
				dir:       h2.ClientToServer,
				processor: cToSProcessor,
				sink:      sinks.ForDirection(h2.ClientToServer),
			}
			cToS = cToSEmitter.adapter
		}
		if sToCProcessor != nil {
			sToCEmitter.adapter = &adapter{
				enabled:   &enabled,
				dir:       h2.ServerToClient,
				processor: sToCProcessor,
				sink:      sinks.ForDirection(h2.ServerToClient),
			}
			sToC = sToCEmitter.adapter
		}
		return cToS, sToC
	}
}

// Processor processes gRPC traffic.
type Processor interface {
	h2.HeaderProcessor
	// Message receives serialized messages.
	Message(data []byte, streamEnded bool) error
}

// dataState represents one of two possible states when consuming gRPC DATA frames.
type dataState uint8

const (
	readingMetadata dataState = iota
	readingMessageData
)

// adapter wraps the Processor interface with an h2.Processor interface. It filters streams that
// are not gRPC and handles decompressing the message data.
type adapter struct {
	enabled *int32

	dir h2.Direction

	processor Processor
	sink      h2.Processor

	encoding Encoding

	// State for the data interpreter.
	buffer     bytes.Buffer
	state      dataState
	compressed bool
	length     uint32
}

func (a *adapter) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	if !a.isEnabled() {
		for _, h := range headers {
			if h.Name == "content-type" && h.Value == "application/grpc" {
				atomic.StoreInt32(a.enabled, 1)
				break
			}
		}
		if !a.isEnabled() {
			return a.sink.Header(headers, streamEnded, priority)
		}
	}

	for _, h := range headers {
		if h.Name == "grpc-encoding" {
			switch h.Value {
			case "identity":
				a.encoding = Identity
			case "gzip":
				a.encoding = Gzip
			case "deflate":
				a.encoding = Deflate
			case "snappy":
				a.encoding = Snappy
			default:
				return fmt.Errorf("unrecognized grpc-encoding %s in %v", h.Value, headers)
			}
		}
	}
	return a.processor.Header(headers, streamEnded, priority)
}

func (a *adapter) Data(data []byte, streamEnded bool) error {
	if !a.isEnabled() {
		return a.sink.Data(data, streamEnded)
	}

	a.buffer.Write(data)

	for {
		switch a.state {
		case readingMetadata:
			if streamEnded && a.buffer.Len() == 0 {
				// gRPC may send empty DATA frames to end a stream.
				if err := a.processor.Message(nil, true); err != nil {
					return err
				}
			}
			if a.buffer.Len() < 5 {
				return nil
			}
			compressed, _ := a.buffer.ReadByte()
			a.compressed = compressed > 0
			if err := binary.Read(&a.buffer, binary.BigEndian, &a.length); err != nil {
				return fmt.Errorf("reading message length: %w", err)
			}
			a.state = readingMessageData
		case readingMessageData:
			if uint32(a.buffer.Len()) < a.length {
				return nil
			}
			data := make([]byte, a.length)
			a.buffer.Read(data)

			if a.compressed {
				switch a.encoding {
				case Identity:
				case Gzip:
					var err error
					data, err = gunzip(data)
					if err != nil {
						return fmt.Errorf("gunzipping data: %w", err)
					}
				case Deflate:
					var err error
					data, err = deflate(data)
					if err != nil {
						return fmt.Errorf("deflating data: %w", err)
					}
				case Snappy:
					var err error
					data, err = ioutil.ReadAll(snappy.NewReader(bytes.NewReader(data)))
					if err != nil {
						return fmt.Errorf("uncompressing snappy: %w", err)
					}
				default:
					panic(fmt.Sprintf("unexpected enocding: %v", a.encoding))
				}
			}
			a.state = readingMetadata

			// Only marks stream ended for the message if there is no data remaining. For ease of
			// implementation, this proxy aligns messages with data frames. This means that if a data
			// frame with stream ended contains multiple messages, the earlier ones should not be
			// marked with stream ended.
			//
			// As explained in https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#data-frames,
			// this reframing is safe because gRPC implementations won't be making any assumptions about
			// the framing.
			if err := a.processor.Message(data, streamEnded && a.buffer.Len() == 0); err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("unexpected state: %v", a.state))
		}
		if a.buffer.Len() == 0 {
			return nil
		}
	}
}

func (a *adapter) Priority(priority http2.PriorityParam) error {
	return a.sink.Priority(priority)
}

func (a *adapter) RSTStream(errCode http2.ErrCode) error {
	return a.sink.RSTStream(errCode)
}

func (a *adapter) PushPromise(promiseID uint32, headers []hpack.HeaderField) error {
	return a.sink.PushPromise(promiseID, headers)
}

func (a *adapter) isEnabled() bool {
	return atomic.LoadInt32(a.enabled) > 0
}

// emitter is a Processor implementation that wraps a h2.Processor instance, forwarding traffic to
// it. It handles recompression of the data.
type emitter struct {
	sink h2.Processor
	// adapter is a reference to the adapter needed to retrieve state.
	adapter *adapter
}

func (e *emitter) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	return e.sink.Header(headers, streamEnded, priority)
}

func (e *emitter) Message(data []byte, streamEnded bool) error {
	// Applies compression to `data` depending on `adapter`'s state.
	if e.adapter.compressed {
		switch e.adapter.encoding {
		case Identity:
		case Gzip:
			var buf bytes.Buffer
			w := gzip.NewWriter(&buf)
			if _, err := w.Write(data); err != nil {
				return fmt.Errorf("gzipping message data: %w", err)
			}
			if err := w.Close(); err != nil {
				return fmt.Errorf("gzipping message data: %w", err)
			}
			data = buf.Bytes()
		case Deflate:
			var buf bytes.Buffer
			w, _ := flate.NewWriter(&buf, -1)
			if _, err := w.Write(data); err != nil {
				return fmt.Errorf("flate compressing message data: %w", err)
			}
			if err := w.Close(); err != nil {
				return fmt.Errorf("flate compressing message data: %w", err)
			}
			data = buf.Bytes()
		case Snappy:
			data = snappy.Encode(nil, data)
		}
	}
	var buf bytes.Buffer
	// Writes the compression status.
	if e.adapter.compressed {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	binary.Write(&buf, binary.BigEndian, uint32(len(data))) // Writes the length of the data.
	buf.Write(data)                                         // Writes the actual data.
	return e.sink.Data(buf.Bytes(), streamEnded)
}

func gunzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func deflate(data []byte) (_ []byte, rerr error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer func() {
		if err := r.Close(); err != nil && rerr != nil {
			rerr = err
		}
	}()
	return ioutil.ReadAll(r)
}
