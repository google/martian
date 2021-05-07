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

// Echo handles TestService.Echo RPC.
func (s *Server) Echo(ctx context.Context, in *tspb.EchoRequest) (*tspb.EchoResponse, error) {
	return &tspb.EchoResponse{
		Payload: in.GetPayload(),
	}, nil
}

func (s *Server) Sum(_ context.Context, in *tspb.SumRequest) (*tspb.SumResponse, error) {
	sum := int32(0)
	for _, v := range in.GetValues() {
		sum += v
	}
	return &tspb.SumResponse{
		Value: sum,
	}, nil
}

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
