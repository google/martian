package marbl

import (
	"io"
	"net/http"

	"github.com/google/martian"
)

type Modifier struct {
	s *Stream
}

func NewModifier(w io.Writer) *Modifier {
	return &Modifier{
		s: NewStream(w),
	}
}

func (m *Modifier) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	return m.s.LogRequest(ctx.ID(), req)
}

func (m *Modifier) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	return m.s.LogResponse(ctx.ID(), res)
}

func (m *Modifier) Close() error {
	return m.s.Close()
}
