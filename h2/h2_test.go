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

package h2_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"sync"
	"testing"

	"github.com/google/martian/v3/h2"
	mgrpc "github.com/google/martian/v3/h2/grpc"
	ht "github.com/google/martian/v3/h2/testing"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/protobuf/proto"

	tspb "github.com/google/martian/v3/h2/testservice"
)

type requestProcessor struct {
	dest     mgrpc.Processor
	requests *[]*tspb.EchoRequest
}

func (p *requestProcessor) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	return p.dest.Header(headers, streamEnded, priority)
}

func (p *requestProcessor) Message(data []byte, streamEnded bool) error {
	msg := &tspb.EchoRequest{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("unmarshalling request: %w", err)
	}
	*p.requests = append(*p.requests, msg)
	return p.dest.Message(data, streamEnded)
}

type responseProcessor struct {
	dest      mgrpc.Processor
	responses *[]*tspb.EchoResponse
}

func (p *responseProcessor) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	return p.dest.Header(headers, streamEnded, priority)
}

func (p *responseProcessor) Message(data []byte, streamEnded bool) error {
	msg := &tspb.EchoResponse{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("unmarshalling response: %w", err)
	}
	*p.responses = append(*p.responses, msg)
	return p.dest.Message(data, streamEnded)
}

func TestEcho(t *testing.T) {
	// This is a basic smoke test. It verifies that the end-to-end flow works and that gRPC messages
	// are observed as expected in processors.
	var requests []*tspb.EchoRequest
	var responses []*tspb.EchoResponse
	fixture, err := ht.New([]h2.StreamProcessorFactory{
		mgrpc.AsStreamProcessorFactory(
			func(_ *url.URL, server, client mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
				return &requestProcessor{server, &requests}, &responseProcessor{client, &responses}
			}),
	})
	if err != nil {
		t.Fatalf("ht.New(...) = %v, want nil", err)
	}
	defer func() {
		if err := fixture.Close(); err != nil {
			t.Fatalf("f.Close() = %v, want nil", err)
		}
	}()

	ctx := context.Background()
	req := &tspb.EchoRequest{
		Payload: "Hello",
	}
	resp, err := fixture.Echo(ctx, req)
	if err != nil {
		t.Fatalf("fixture.Echo(...) = _, %v, want _, nil", err)
	}
	if got, want := resp.GetPayload(), req.GetPayload(); got != want {
		t.Errorf("resp.GetPayload() = %s, want = %s", got, want)
	}

	// Verifies the captured requests and responses.
	if got := len(requests); got != 1 {
		t.Fatalf("len(requests) = %d, want 1", got)
	}
	if got, want := requests[0].GetPayload(), req.GetPayload(); got != want {
		t.Errorf("requests[0].GetPayload() = %s, want = %s", got, want)
	}
	if got := len(responses); got != 1 {
		t.Fatalf("len(requests) = %d, want 1", got)
	}
	if got, want := responses[0].GetPayload(), req.GetPayload(); got != want {
		t.Errorf("responses[0].GetPayload() = %s, want = %s", got, want)
	}
}

type requestEditor struct {
	dest mgrpc.Processor
}

func (p *requestEditor) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	return p.dest.Header(headers, streamEnded, priority)
}

func (p *requestEditor) Message(_ []byte, streamEnded bool) error {
	msg := &tspb.EchoRequest{
		Payload: "Goodbye",
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshalling request: %w", err)
	}
	return p.dest.Message(data, streamEnded)
}

func TestRequestEditor(t *testing.T) {
	// This test inserts a request modifier that changes the payload from "Hello" to "Goodbye".
	fixture, err := ht.New([]h2.StreamProcessorFactory{
		mgrpc.AsStreamProcessorFactory(
			func(_ *url.URL, server, client mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
				return &requestEditor{server}, nil
			}),
	})
	if err != nil {
		t.Fatalf("ht.New(...) = %v, want nil", err)
	}
	defer func() {
		if err := fixture.Close(); err != nil {
			t.Fatalf("f.Close() = %v, want nil", err)
		}
	}()

	ctx := context.Background()
	req := &tspb.EchoRequest{
		Payload: "Hello",
	}
	resp, err := fixture.Echo(ctx, req)
	if err != nil {
		t.Fatalf("fixture.Echo(...) = _, %v, want _, nil", err)
	}
	if got, want := resp.GetPayload(), "Goodbye"; got != want {
		t.Errorf("resp.GetPayload() = %s, want = %s", got, want)
	}
}

type plusOne struct {
	dest mgrpc.Processor
}

func (p *plusOne) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	return p.dest.Header(headers, streamEnded, priority)
}

func (p *plusOne) Message(data []byte, streamEnded bool) error {
	msg := &tspb.SumRequest{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("unmarshalling request: %w", err)
	}
	msg.Values = append(msg.Values, 1)

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshalling request: %w", err)
	}
	return p.dest.Message(data, streamEnded)
}

func TestProcessorChaining(t *testing.T) {
	// This test constructs a chain of processors and checks that the effects are correctly applied
	// at the result.
	fixture, err := ht.New([]h2.StreamProcessorFactory{
		mgrpc.AsStreamProcessorFactory(
			func(_ *url.URL, server, client mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
				return &plusOne{server}, nil
			}),
		mgrpc.AsStreamProcessorFactory(
			func(_ *url.URL, server, client mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
				return &plusOne{server}, nil
			}),
		mgrpc.AsStreamProcessorFactory(
			func(_ *url.URL, server, client mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
				return &plusOne{server}, nil
			}),
	})
	if err != nil {
		t.Fatalf("ht.New(...) = %v, want nil", err)
	}
	defer func() {
		if err := fixture.Close(); err != nil {
			t.Fatalf("f.Close() = %v, want nil", err)
		}
	}()

	ctx := context.Background()
	req := &tspb.SumRequest{
		Values: []int32{5},
	}
	resp, err := fixture.Sum(ctx, req)
	if err != nil {
		t.Fatalf("fixture.Sum(...) = _, %v, want _, nil", err)
	}
	if got, want := resp.GetValue(), int32(8); got != want {
		t.Errorf("resp.GetValue() = %d, want = %d", got, want)
	}
}

type headerCapture struct {
	dest    mgrpc.Processor
	headers *[][]hpack.HeaderField
}

func (h *headerCapture) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	c := make([]hpack.HeaderField, len(headers))
	copy(c, headers)
	*h.headers = append(*h.headers, c)
	return h.dest.Header(headers, streamEnded, priority)
}

func (h *headerCapture) Message(data []byte, streamEnded bool) error {
	return h.dest.Message(data, streamEnded)
}

func TestLargeEcho(t *testing.T) {
	// Sends a >128KB payload through the proxy. Since the standard gRPC frame size is only 16KB,
	// this exercises frame merging, splitting and flow control code.
	payload := make([]byte, 128*1024)
	rand.Read(payload)
	req := &tspb.EchoRequest{
		Payload: base64.StdEncoding.EncodeToString(payload),
	}

	// This test also covers using gzip compression. Ideally, we would test more compression types
	// but the golang gRPC implementation only provides a gzip compressor.
	tests := []struct {
		name           string
		useCompression bool
	}{
		{"RawData", false},
		{"Gzip", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var cToSHeaders, sToCHeaders [][]hpack.HeaderField
			fixture, err := ht.New([]h2.StreamProcessorFactory{
				mgrpc.AsStreamProcessorFactory(
					func(_ *url.URL, server, client mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
						return &headerCapture{server, &cToSHeaders}, &headerCapture{client, &sToCHeaders}
					}),
			})
			if err != nil {
				t.Fatalf("ht.New(...) = %v, want nil", err)
			}
			defer func() {
				if err := fixture.Close(); err != nil {
					t.Fatalf("f.Close() = %v, want nil", err)
				}
			}()

			ctx := context.Background()
			var resp *tspb.EchoResponse
			if tc.useCompression {
				resp, err = fixture.Echo(ctx, req, grpc.UseCompressor(gzip.Name))
			} else {
				resp, err = fixture.Echo(ctx, req)
			}
			if err != nil {
				t.Fatalf("fixture.Echo(...) = _, %v, want _, nil", err)
			}
			if got, want := resp.GetPayload(), req.GetPayload(); got != want {
				t.Errorf("resp.GetPayload() = %s, want = %s", got, want)
			}
			// Verifies that grpc-encoding=gzip is present in the first headers on the stream when
			// compression is active.
			for _, headers := range [][]hpack.HeaderField{cToSHeaders[0], sToCHeaders[0]} {
				foundGRPCEncoding := false
				for _, h := range headers {
					if h.Name == "grpc-encoding" {
						foundGRPCEncoding = true
						if got, want := h.Value, "gzip"; got != want {
							t.Errorf("h.Value = %s, want %s", got, want)
						}
					}
				}
				if got, want := foundGRPCEncoding, tc.useCompression; got != want {
					t.Errorf("foundGRPCEncoding = %t, want %t", got, want)
				}
			}
		})
	}
}

type noopProcessor struct {
	sink mgrpc.Processor
}

func (p *noopProcessor) Header(
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	return p.sink.Header(headers, streamEnded, priority)
}

func (p *noopProcessor) Message(data []byte, streamEnded bool) error {
	return p.sink.Message(data, streamEnded)
}

func TestStream(t *testing.T) {
	tests := []struct {
		name    string
		factory h2.StreamProcessorFactory
	}{
		{
			"NilH2Processor",
			func(_ *url.URL, _ *h2.Processors) (h2.Processor, h2.Processor) {
				return nil, nil
			},
		},
		{
			// This differs from NilH2Processor only in how mgrpc.AsStreamProcessorFactory handles nil
			// grpc.Processor values. It should end up processing exactly the same as
			// h2.StreamProcessorFactory afterwards.
			"NilGRPCProcessor",
			mgrpc.AsStreamProcessorFactory(
				func(_ *url.URL, _, _ mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
					return nil, nil
				}),
		},
		{
			// This differs from NilGRPCProcessor in that NilGRPCProcessor ends up behaving like
			// NilH2Processor and no gRPC processing takes place. NoopProcessor causes the frames to
			// be processed as gRPC.
			"NoopGRPCProcessor",
			mgrpc.AsStreamProcessorFactory(
				func(_ *url.URL, server, client mgrpc.Processor) (mgrpc.Processor, mgrpc.Processor) {
					return &noopProcessor{server}, &noopProcessor{client}
				}),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture, err := ht.New([]h2.StreamProcessorFactory{tc.factory})
			if err != nil {
				t.Fatalf("ht.New(...) = %v, want nil", err)
			}
			defer func() {
				if err := fixture.Close(); err != nil {
					t.Fatalf("f.Close() = %v, want nil", err)
				}
			}()
			ctx := context.Background()
			stream, err := fixture.DoubleEcho(ctx)
			if err != nil {
				t.Fatalf("fixture.DoubleEcho(ctx) = _, %v, want _, nil", err)
			}

			var received []*tspb.EchoResponse

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					resp, err := stream.Recv()
					if err == io.EOF {
						return
					}
					if err != nil {
						t.Errorf("stream.Recv() = %v, want nil", err)
						return
					}
					received = append(received, resp)
				}
			}()

			var sent []*tspb.EchoRequest
			for i := 0; i < 5; i++ {
				payload := make([]byte, 20*1024)
				rand.Read(payload)
				req := &tspb.EchoRequest{
					Payload: base64.StdEncoding.EncodeToString(payload),
				}
				if err := stream.Send(req); err != nil {
					t.Fatalf("stream.Send(req) = %v, want nil", err)
				}
				sent = append(sent, req)
			}
			if err := stream.CloseSend(); err != nil {
				t.Fatalf("stream.CloseSend() = %v, want nil", err)
			}
			wg.Wait()

			for i, req := range sent {
				want := req.GetPayload()
				if got := received[2*i].GetPayload(); got != want {
					t.Errorf("received[2*i].GetPayload() = %s, want %s", got, want)
				}
				if got := received[2*i+1].GetPayload(); got != want {
					t.Errorf("received[2*i+1].GetPayload() = %s, want %s", got, want)
				}
			}
		})
	}
}
