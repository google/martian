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

package martian

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/martian/proxyutil"
)

type hijackResponseRecorder struct {
	Code      int
	Flushed   bool
	HeaderMap http.Header
	Body      *bytes.Buffer

	conn                  net.Conn
	wroteHeader, hijacked bool
}

func newHijackRecorder(conn net.Conn) *hijackResponseRecorder {
	return &hijackResponseRecorder{
		conn:      conn,
		Body:      new(bytes.Buffer),
		HeaderMap: http.Header{},
	}
}

func (rw *hijackResponseRecorder) Flush() {
	if rw.hijacked {
		return
	}

	rw.Flushed = true
}

func (rw *hijackResponseRecorder) Header() http.Header {
	if rw.hijacked {
		return nil
	}

	return rw.HeaderMap
}

func (rw *hijackResponseRecorder) Write(b []byte) (n int, err error) {
	if rw.hijacked {
		return 0, fmt.Errorf("connection hijacked")
	}

	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}

	if _, err := rw.Body.Write(b); err != nil {
		return 0, err
	}
	return rw.conn.Write(b)
}

func (rw *hijackResponseRecorder) WriteHeader(code int) {
	if rw.hijacked || rw.wroteHeader {
		return
	}

	rw.wroteHeader = true
	rw.Code = code

	status := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, http.StatusText(code))
	rw.conn.Write([]byte(status))
	rw.HeaderMap.Write(rw.conn)
	rw.conn.Write([]byte("\r\n"))
}

func (rw *hijackResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw.hijacked = true

	br := bufio.NewReader(rw.conn)
	bw := bufio.NewWriter(rw.conn)
	return rw.conn, bufio.NewReadWriter(br, bw), nil
}

type timeoutPipe struct {
	net.Conn
	readTimeout  time.Time
	writeTimeout time.Time
}

type pipeNetError struct {
	timeout, temporary bool
}

func (e *pipeNetError) Error() string {
	return "pipe error"
}

func (e *pipeNetError) Temporary() bool {
	return e.temporary
}

func (e *pipeNetError) Timeout() bool {
	return e.timeout
}

func pipeWithTimeout() (*timeoutPipe, *timeoutPipe) {
	rc, wc := net.Pipe()

	trc := &timeoutPipe{
		Conn:         rc,
		readTimeout:  time.Now().Add(3 * time.Minute),
		writeTimeout: time.Now().Add(3 * time.Minute),
	}

	twc := &timeoutPipe{
		Conn:         wc,
		readTimeout:  time.Now().Add(3 * time.Minute),
		writeTimeout: time.Now().Add(3 * time.Minute),
	}
	return trc, twc
}

func (p *timeoutPipe) SetDeadline(t time.Time) error {
	p.SetReadDeadline(t)
	p.SetWriteDeadline(t)
	return nil
}

func (p *timeoutPipe) SetReadDeadline(t time.Time) error {
	p.readTimeout = t
	return nil
}

func (p *timeoutPipe) SetWriteDeadline(t time.Time) error {
	p.writeTimeout = t
	return nil
}

func (p *timeoutPipe) Read(b []byte) (int, error) {
	type connRead struct {
		n   int
		err error
	}

	rc := make(chan connRead, 1)

	go func() {
		n, err := p.Conn.Read(b)
		rc <- connRead{
			n:   n,
			err: err,
		}
	}()

	d := p.readTimeout.Sub(time.Now())
	if d <= 0 {
		return 0, &pipeNetError{true, false}
	}

	select {
	case cr := <-rc:
		return cr.n, cr.err
	case <-time.After(d):
		return 0, &pipeNetError{true, false}
	}
}

func (p *timeoutPipe) Write(b []byte) (n int, err error) {
	type connWrite struct {
		n   int
		err error
	}
	wc := make(chan connWrite, 1)

	go func() {
		n, err := p.Conn.Write(b)
		wc <- connWrite{
			n:   n,
			err: err,
		}
	}()

	d := p.writeTimeout.Sub(time.Now())
	if d <= 0 {
		return 0, &pipeNetError{true, false}
	}

	select {
	case cw := <-wc:
		return cw.n, cw.err
	case <-time.After(d):
		return 0, &pipeNetError{true, false}
	}
}

func init() {
	http.DefaultTransport = RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return proxyutil.NewResponse(200, nil, req), nil
	})
}

func tlsClient(conn net.Conn, ca *x509.Certificate, server string) net.Conn {
	pool := x509.NewCertPool()
	pool.AddCert(ca)

	return tls.Client(conn, &tls.Config{
		ServerName: server,
		RootCAs:    pool,
	})
}

func TestModifyRequest(t *testing.T) {
	p := NewProxy(mitm)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q, ..., nil): got %v, want no error", "GET", err)
	}

	if err := p.ModifyRequest(NewContext(), req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	modifierRun := false
	f := func(*Context, *http.Request) error {
		modifierRun = true
		return nil
	}
	p.SetRequestModifier(RequestModifierFunc(f))

	if err := p.ModifyRequest(NewContext(), req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if !modifierRun {
		t.Error("modifierRun: got false, want true")
	}
}

func TestModifyResponse(t *testing.T) {
	p := NewProxy(mitm)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q, ..., nil): got %v, want no error", "GET", err)
	}
	res := proxyutil.NewResponse(200, nil, req)

	if err := p.ModifyResponse(NewContext(), res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	modifierRun := false
	f := func(*Context, *http.Response) error {
		modifierRun = true
		return nil
	}
	p.SetResponseModifier(ResponseModifierFunc(f))

	if err := p.ModifyResponse(NewContext(), res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}
	if !modifierRun {
		t.Error("modifierRun: got false, want true")
	}
}

func TestServeHTTPHijackConversionError(t *testing.T) {
	p := NewProxy(mitm)
	// httptest.ResponseRecorder is not an http.Hijacker.
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	p.ServeHTTP(rw, req)

	if got, want := rw.Code, 500; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}
	if got, want := rw.Body.String(), "error unsupported http.Hijacker\n"; got != want {
		t.Errorf("rw.Body: got %q, want %q", got, want)
	}
}

func TestServeHTTPModifyConnectRequestError(t *testing.T) {
	p := NewProxy(mitm)
	p.SetConnectRequestModifier(RequestModifierFunc(
		func(*Context, *http.Request) error {
			return fmt.Errorf("modifier error")
		}))

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 400; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	if want := []byte("modifier error"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}

func TestServeHTTPModifyConnectResponseError(t *testing.T) {
	p := NewProxy(mitm)
	p.SetConnectResponseModifier(ResponseModifierFunc(
		func(*Context, *http.Response) error {
			return fmt.Errorf("modifier error")
		}))

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 400; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	if want := []byte("modifier error"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}

func TestServeHTTPReadRequestError(t *testing.T) {
	p := NewProxy(mitm)
	// Shorten the timeout to force a ReadRequest error.
	p.Timeout = time.Second

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	tlsConn.Write([]byte("INVALID /invalid NOTHTTP/1.1\r\n"))

	if _, err = http.ReadResponse(bufio.NewReader(tlsConn), nil); err != io.ErrUnexpectedEOF {
		t.Fatalf("http.ReadResponse(): got %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestServeHTTPModifyRequestError(t *testing.T) {
	p := NewProxy(mitm)
	f := func(*Context, *http.Request) error {
		return fmt.Errorf("modifier error")
	}
	p.SetRequestModifier(RequestModifierFunc(f))

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	req, err = http.NewRequest("GET", "https://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no erro", err)
	}
	if err := req.Write(tlsConn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err = http.ReadResponse(bufio.NewReader(tlsConn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 400; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	if want := []byte("modifier error"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}

func TestServeHTTPRoundTripError(t *testing.T) {
	p := NewProxy(mitm)
	p.RoundTripper = RoundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("round trip error")
	})

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	req, err = http.NewRequest("GET", "https://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no erro", err)
	}
	if err := req.Write(tlsConn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err = http.ReadResponse(bufio.NewReader(tlsConn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 502; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	if want := []byte("round trip error"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}

func TestServeHTTPModifyResponseError(t *testing.T) {
	p := NewProxy(mitm)
	f := func(*Context, *http.Response) error {
		return fmt.Errorf("modifier error")
	}
	p.SetResponseModifier(ResponseModifierFunc(f))

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	req, err = http.NewRequest("GET", "https://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no erro", err)
	}
	if err := req.Write(tlsConn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err = http.ReadResponse(bufio.NewReader(tlsConn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 400; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}
	if want := []byte("modifier error"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}

func TestServeHTTPSkipRoundTrip(t *testing.T) {
	p := NewProxy(mitm)
	f := func(ctx *Context, _ *http.Request) error {
		ctx.SkipRoundTrip = true
		return nil
	}
	p.SetRequestModifier(RequestModifierFunc(f))

	p.RoundTripper = RoundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("RoundTrip(): got called, want skipped")
		return nil, nil
	})

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	req, err = http.NewRequest("GET", "https://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Connection", "close")

	if err := req.Write(tlsConn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err = http.ReadResponse(bufio.NewReader(tlsConn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
}

func TestServeHTTPBuildsValidRequest(t *testing.T) {
	p := NewProxy(mitm)
	p.RoundTripper = RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.URL.Scheme, "https"; got != want {
			t.Errorf("req.URL.Scheme: got %q, want %q", got, want)
		}
		if got, want := req.URL.Host, "www.example.com"; got != want {
			t.Errorf("req.URL.Host: got %q, want %q", got, want)
		}
		if req.RemoteAddr == "" {
			t.Error("req.RemoteAddr: got empty, want addr")
		}

		return proxyutil.NewResponse(201, nil, req), nil
	})

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	req, err = http.NewRequest("GET", "https://www.example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no erro", err)
	}
	req.Header.Set("Connection", "close")

	if err := req.Write(tlsConn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err = http.ReadResponse(bufio.NewReader(tlsConn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 201; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
}

func TestServeHTTPRunsModifiers(t *testing.T) {
	p := NewProxy(mitm)
	modsRun := []string{}

	p.SetConnectRequestModifier(RequestModifierFunc(
		func(*Context, *http.Request) error {
			modsRun = append(modsRun, "creqmod")
			return nil
		}))

	p.SetRequestModifier(RequestModifierFunc(
		func(*Context, *http.Request) error {
			modsRun = append(modsRun, "reqmod")
			return nil
		}))

	p.SetConnectResponseModifier(ResponseModifierFunc(
		func(*Context, *http.Response) error {
			modsRun = append(modsRun, "cresmod")
			return nil
		}))

	p.SetResponseModifier(ResponseModifierFunc(
		func(*Context, *http.Response) error {
			modsRun = append(modsRun, "resmod")
			return nil
		}))

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	req, err = http.NewRequest("GET", "https://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no erro", err)
	}
	req.Header.Set("Connection", "close")

	if err := req.Write(tlsConn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err = http.ReadResponse(bufio.NewReader(tlsConn), nil)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}

	if got, want := modsRun, []string{"creqmod", "cresmod", "reqmod", "resmod"}; !reflect.DeepEqual(got, want) {
		t.Errorf("modsRun: got %v, want %v", got, want)
	}
}

func TestServeHTTPTimeout(t *testing.T) {
	p := NewProxy(mitm)

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")

	// Set timeout in the past to force timeout error.
	tlsConn.SetDeadline(time.Now().Add(-time.Second))

	req, err = http.NewRequest("GET", "https://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := req.Write(tlsConn); err == nil {
		t.Fatal("req.Write(): got nil, want error")
	}
}

func TestServeHTTPKeepAlive(t *testing.T) {
	p := NewProxy(mitm)

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")

	tt := []struct {
		closing bool
	}{
		{false},
		{false},
		{true},
	}
	for _, tc := range tt {
		req, err = http.NewRequest("GET", "https://www.example.com/", nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		// Close the connection on the last request.
		if tc.closing {
			req.Header.Set("Connection", "close")
		}

		if err := req.Write(tlsConn); err != nil {
			t.Fatalf("req.Write(): got %v, want no error", err)
		}

		res, err = http.ReadResponse(bufio.NewReader(tlsConn), nil)
		if err != nil {
			t.Fatalf("http.ReadResponse(): got %v, want no error", err)
		}
		res.Body.Close()

		if got, want := res.StatusCode, 200; got != want {
			t.Errorf("res.StatusCode: got %d, want %d", got, want)
		}
		if tc.closing && !res.Close {
			t.Error("res.Close: got false, want true")
		}
	}
}

func TestServeHTTPChunkedBody(t *testing.T) {
	p := NewProxy(mitm)
	p.RoundTripper = RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		res := proxyutil.NewResponse(200, strings.NewReader(`chunked body`), req)
		res.Header.Set("Transfer-Encoding", "chunked")
		res.TransferEncoding = []string{"chunked"}

		return res, nil
	})

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)
	req, err := http.NewRequest("CONNECT", "//www.example.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	res.Body.Close()

	tlsConn := tlsClient(rc, p.mitm.Authority, "www.example.com")
	req, err = http.NewRequest("GET", "https://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Connection", "close")

	if err := req.Write(tlsConn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	got, err := ioutil.ReadAll(tlsConn)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}

	// The reason for the initial single character chunk is related to the
	// use of io.MultiReader in the http.Response#Write function and the
	// implementation of io.MultiReader#Read.
	want := "1\r\nc\r\nb\r\nhunked body\r\n0\r\n\r\n"
	if !bytes.Contains(got, []byte(want)) {
		t.Errorf("tlsConn: got %q, want %q", got, want)
	}
}

func TestShouldCloseAfterReply(t *testing.T) {
	tt := []struct {
		values []string
		want   bool
	}{
		{[]string{""}, false},
		{[]string{"X-Hop-By-Hop", "Close"}, true},
		{[]string{"X-HBH, close", "X-Hop-By-Hop"}, true},
		{[]string{"X-close"}, false},
	}

	for _, tc := range tt {
		header := http.Header{"Connection": tc.values}

		if got := shouldCloseAfterReply(header); got != tc.want {
			t.Errorf("shouldCloseAfterReply(%v): got %t, want %t", header, got, tc.want)
		}
	}
}

func TestServeHTTPConnectRequestWithoutMITM(t *testing.T) {
	p := NewProxy(nil)
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret!"))
	}
	server := httptest.NewTLSServer(http.HandlerFunc(f))
	defer server.Close()

	rc, wc := pipeWithTimeout()
	defer rc.Close()
	defer wc.Close()

	rw := newHijackRecorder(wc)

	req, err := http.NewRequest("CONNECT", server.URL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q, %q, nil): got %v, want no error", "CONNECT", server.URL, err)
	}

	go p.ServeHTTP(rw, req)

	res, err := http.ReadResponse(bufio.NewReader(rc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}

	req, err = http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q, %q, nil): got %v, want no error", "GET", server.URL, err)
	}
	req.Header.Set("Connection", "close")

	tlsrc := tls.Client(rc, &tls.Config{
		InsecureSkipVerify: true,
	})

	go req.Write(tlsrc)

	res, err = http.ReadResponse(bufio.NewReader(tlsrc), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(res.Body): got %v, want no error", err)
	}
	res.Body.Close()

	if want := []byte("secret!"); !bytes.Equal(got, want) {
		t.Errorf("res.Body: got %q, want %q", got, want)
	}
}
