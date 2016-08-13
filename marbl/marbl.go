package marbl

import (
	"io"
	"net/http"
	"strconv"
	"time"
	"sync/atomic"

	"github.com/google/martian/proxyutil"
)

// Frame Header
// FrameType	 uint8
// MessageType uint8
// ID					 [8]byte
// Payload		 HeaderFrame/DataFrame

// Header Frame
// NameLen  uint32
// ValueLen uint32
// Name	    variable
// Value    variable

// Data Frame
// Index uint32
// Terminal uint8
// Len  uint32
// Data variable

type MessageType uint8
type FrameType uint8

const (
	Unknown  MessageType = 0x0
	Request  MessageType = 0x1
	Response MessageType = 0x2
)

const (
	UnknownFrame FrameType = 0x0
	HeaderFrame  FrameType = 0x1
	DataFrame    FrameType = 0x2
)

type Stream struct {
	w      io.Writer
	framec chan []byte
	closec chan struct{}
}

func NewStream(w io.Writer) *Stream {
	s := &Stream{
		w:      w,
		framec: make(chan []byte),
		closec: make(chan struct{}),
	}

	go s.loop()

	return s
}

func (s *Stream) loop() {
	for {
		select {
		case f := <-s.framec:
			_, err := s.w.Write(f)
			if err != nil {
				// log the error.
			}
		case <-s.closec:
			return
		}
	}
}

func (s *Stream) Close() error {
	s.closec <- struct{}{}
	close(s.closec)

	return nil
}

func newFrame(id string, ft FrameType, mt MessageType, plen uint32) []byte {
	// FrameType:   1 byte
	// MessageType: 1 byte
	// ID:			    8 bytes
	// Payload:     plen bytes

	f := make([]byte, 0, 10+plen)
	f = append(f, byte(ft), byte(mt))
	f = append(f, id[:8]...)

	return f
}

func (s *Stream) sendHeader(id string, mt MessageType, key, value string) {
	kl := uint32(len(key))
	vl := uint32(len(value))

	f := newFrame(id, HeaderFrame, mt, 64+kl+vl)
	f = append(f, byte(kl>>24), byte(kl>>16), byte(kl>>8), byte(kl))
	f = append(f, byte(vl>>24), byte(vl>>16), byte(vl>>8), byte(vl))
	f = append(f, key[:kl]...)
	f = append(f, value[:vl]...)

	s.framec <- f
}

func (s *Stream) sendData(id string, mt MessageType, i uint32, terminal bool, b []byte) {
	bl := uint32(len(b))

	var ti uint8
	if terminal {
		ti = 1
	}

	f := newFrame(id, DataFrame, mt, 40+bl)
	f = append(f, byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
	f = append(f, byte(ti>>8), byte(ti))
	f = append(f, byte(bl>>24), byte(bl>>16), byte(bl>>8), byte(bl))
	f = append(f, b...)

	s.framec <- f
}

func (s *Stream) LogRequest(id string, req *http.Request) error {
	s.sendHeader(id, Request, ":method", req.Method)
	s.sendHeader(id, Request, ":scheme", req.URL.Scheme)
	s.sendHeader(id, Request, ":authority", req.URL.Host)
	s.sendHeader(id, Request, ":path", req.URL.Path)
	s.sendHeader(id, Request, ":query", req.URL.RawQuery)
	s.sendHeader(id, Request, ":proto", req.Proto)
	s.sendHeader(id, Request, ":remote", req.RemoteAddr)
	ts := strconv.FormatInt(time.Now().UnixNano() / 1000 / 1000, 10)
	s.sendHeader(id, Request, ":timestamp", ts)

	h := proxyutil.RequestHeader(req)

	for k, vs := range h.Map() {
		for _, v := range vs {
			s.sendHeader(id, Request, k, v)
		}
	}

	req.Body = &bodyLogger{
		s:    s,
		id:   id,
		mt:   Request,
		body: req.Body,
	}

	return nil
}

func (s *Stream) LogResponse(id string, res *http.Response) error {
	s.sendHeader(id, Response, ":proto", res.Proto)
	s.sendHeader(id, Response, ":status", strconv.Itoa(res.StatusCode))
	s.sendHeader(id, Response, ":reason", res.Status)
	ts := strconv.FormatInt(time.Now().UnixNano() / 1000 / 1000, 10)
	s.sendHeader(id, Response, ":timestamp", ts)

	h := proxyutil.ResponseHeader(res)

	for k, vs := range h.Map() {
		for _, v := range vs {
			s.sendHeader(id, Response, k, v)
		}
	}

	res.Body = &bodyLogger{
		s:    s,
		id:   id,
		mt:   Response,
		body: res.Body,
	}

	return nil
}

type bodyLogger struct {
	index uint32 // atomic
	s     *Stream
	id    string
	mt    MessageType
	body  io.ReadCloser
}

func (bl *bodyLogger) Read(b []byte) (int, error) {
	var terminal bool

	n, err := bl.body.Read(b)
	if err == io.EOF {
		terminal = true
	}

	bl.s.sendData(bl.id, bl.mt, atomic.AddUint32(&bl.index, 1)-1, terminal, b)

	return n, err
}

func (bl *bodyLogger) Close() error {
	return bl.body.Close()
}
