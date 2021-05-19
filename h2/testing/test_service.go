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

package testing

import (
	"context"
	"io"

	tspb "github.com/google/martian/v3/h2/testservice"
)

// Server is a testing gRPC server.
type Server struct {
	tspb.UnimplementedTestServiceServer
}

// Echo handles TestService.Echo RPCs.
func (s *Server) Echo(ctx context.Context, in *tspb.EchoRequest) (*tspb.EchoResponse, error) {
	return &tspb.EchoResponse{
		Payload: in.GetPayload(),
	}, nil
}

// Sum handles TestService.Sum RPCs.
func (s *Server) Sum(_ context.Context, in *tspb.SumRequest) (*tspb.SumResponse, error) {
	sum := int32(0)
	for _, v := range in.GetValues() {
		sum += v
	}
	return &tspb.SumResponse{
		Value: sum,
	}, nil
}

// DoubleEcho handles TestService.DoubleEcho RPCs.
func (s *Server) DoubleEcho(stream tspb.TestService_DoubleEchoServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		resp := &tspb.EchoResponse{
			Payload: req.GetPayload(),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}
