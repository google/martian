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
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/martian/mitm"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/session"
)

var (
	closeConn = errors.New("closing connection")

	ctxmu sync.RWMutex
	ctxs  = make(map[*http.Request]*session.Context)
)

// SetContext associates the context with request.
func SetContext(req *http.Request, ctx *session.Context) {
	ctxmu.Lock()
	defer ctxmu.Unlock()

	ctxs[req] = ctx
}

// RemoveContext removes the context for a given request.
func RemoveContext(req *http.Request) {
	ctxmu.Lock()
	defer ctxmu.Unlock()

	delete(ctxs, req)
}

// Context returns the associated context for the request.
func Context(req *http.Request) *session.Context {
	ctxmu.RLock()
	defer ctxmu.RUnlock()

	return ctxs[req]
}

func isCloseableError(err error) bool {
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return true
	}

	switch err {
	case io.EOF, io.ErrClosedPipe, closeConn:
		return true
	}

	return false
}

// Proxy is an HTTP proxy with support for TLS MITM and customizable behavior.
type Proxy struct {
	roundTripper http.RoundTripper
	timeout      time.Duration
	mitm         *mitm.Config
	proxyURL     *url.URL
	conns        *sync.WaitGroup
	closing      int32 // atomic

	reqmod RequestModifier
	resmod ResponseModifier
}

// Option is a configurable proxy setting.
type Option func(p *Proxy)

// NewProxy returns a new HTTP proxy.
func NewProxy() *Proxy {
	nm := Noop("martian")

	return &Proxy{
		roundTripper: http.DefaultTransport,
		timeout:      5 * time.Minute,
		conns:        &sync.WaitGroup{},
		reqmod:       nm,
		resmod:       nm,
	}
}

// Option sets the given options for the proxy.
func (p *Proxy) Option(opts ...Option) {
	for _, opt := range opts {
		opt(p)
	}
}

// RoundTripper sets the http.RoundTripper of the proxy.
func RoundTripper(rt http.RoundTripper) Option {
	return func(p *Proxy) {
		p.roundTripper = rt
	}
}

// Timeout sets the request timeout of the proxy.
func Timeout(timeout time.Duration) Option {
	return func(p *Proxy) {
		p.timeout = timeout
	}
}

// MITM sets the config to use for MITMing of CONNECT requests.
func MITM(config *mitm.Config) Option {
	return func(p *Proxy) {
		p.mitm = config
	}
}

// DownstreamProxy sets the proxy that receives requests from the upstream
// proxy.
func DownstreamProxy(proxyURL *url.URL) Option {
	return func(p *Proxy) {
		p.proxyURL = proxyURL
	}
}

// Close sets the proxying to the closing state and waits for all connections
// to resolve.
func (p *Proxy) Close() {
	atomic.StoreInt32(&p.closing, 1)
	p.conns.Wait()
}

// Closing returns whether the proxy is in the closing state.
func (p *Proxy) Closing() bool {
	return atomic.LoadInt32(&p.closing) == 1
}

// SetRequestModifier sets the request modifier.
func (p *Proxy) SetRequestModifier(reqmod RequestModifier) {
	p.reqmod = reqmod
}

// SetResponseModifier sets the response modifier.
func (p *Proxy) SetResponseModifier(resmod ResponseModifier) {
	p.resmod = resmod
}

// Serve accepts connections from the listener and handles the requests.
func (p *Proxy) Serve(l net.Listener) error {
	defer l.Close()

	var delay time.Duration
	for {
		if p.Closing() {
			return nil
		}

		conn, err := l.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if max := time.Second; delay > max {
					delay = max
				}

				Debugf("martian: temporary error on accept: %v", err)
				time.Sleep(delay)
				continue
			}

			Errorf("martian: failed to accept: %v", err)
			return err
		}
		delay = 0
		Debugf("martian: accepted connection from %s", conn.RemoteAddr())

		if tconn, ok := conn.(*net.TCPConn); ok {
			tconn.SetKeepAlive(true)
			tconn.SetKeepAlivePeriod(3 * time.Minute)
		}

		go p.handleLoop(conn)
	}
}

func (p *Proxy) handleLoop(conn net.Conn) {
	p.conns.Add(1)
	defer p.conns.Done()
	defer conn.Close()

	ctx := session.NewContext()
	brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		deadline := time.Now().Add(p.timeout)
		if err := conn.SetDeadline(deadline); err != nil {
			Errorf("martian: failed to set connection deadline: %v", err)
			return
		}

		if err := p.handle(ctx, conn, brw); isCloseableError(err) {
			Infof("martian: closing connection: %v", conn.RemoteAddr())
			return
		}
	}
}

func (p *Proxy) handle(ctx *session.Context, conn net.Conn, brw *bufio.ReadWriter) error {
	Debugf("martian: waiting for request: %v", conn.RemoteAddr())

	req, err := http.ReadRequest(brw.Reader)
	if err != nil {
		if !isCloseableError(err) {
			Errorf("martian: failed to read request: %v", err)
		}

		// Failed to read request, close the connection.
		return closeConn
	}

	req.URL.Scheme = "http"
	if ctx.IsSecure() {
		Debugf("martian: forcing HTTPS inside secure session")
		req.URL.Scheme = "https"
	}

	req.RemoteAddr = conn.RemoteAddr().String()
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	if tlsconn, ok := conn.(*tls.Conn); ok {
		cs := tlsconn.ConnectionState()
		req.TLS = &cs
	}

	Debugf("martian: received request: %s", req.URL)

	ctx, err = session.FromContext(ctx)
	if err != nil {
		Errorf("martian: failed to derive context: %v", err)
		return err
	}

	SetContext(req, ctx)
	defer RemoveContext(req)

	if req.Method == "CONNECT" {
		if err := p.reqmod.ModifyRequest(req); err != nil {
			Errorf("martian: error modifying CONNECT request: %v", err)
			proxyutil.Warning(req.Header, err)
		}

		if p.mitm != nil {
			Debugf("martian: attempting MITM for connection: %s", req.Host)
			res := proxyutil.NewResponse(200, nil, req)

			if err := p.resmod.ModifyResponse(res); err != nil {
				Errorf("martian: error modifying CONNECT response: %v", err)
				proxyutil.Warning(res.Header, err)
			}

			if err := res.Write(brw); err != nil {
				Errorf("martian: failed to write CONNECT response: %v", err)
				return err
			}
			brw.Flush()

			Debugf("martian: completed MITM for connection: %s", req.Host)
			ctx.MarkSecure()

			tlsconn := tls.Server(conn, p.mitm.TLSForHost(req.Host))
			brw.Writer.Reset(tlsconn)
			brw.Reader.Reset(tlsconn)

			return p.handle(ctx, tlsconn, brw)
		}

		Debugf("martian: attempting to establish CONNECT tunnel: %s", req.URL.Host)
		res, dconn, cerr := p.connect(req)
		if cerr != nil {
			Errorf("martian: failed to CONNECT: %v", err)
			res = proxyutil.NewResponse(502, nil, req)
			proxyutil.Warning(res.Header, cerr)
		}

		if err := p.resmod.ModifyResponse(res); err != nil {
			Errorf("martian: error modifying CONNECT response: %v", err)
			proxyutil.Warning(res.Header, err)
		}

		if err := res.Write(brw); err != nil {
			Errorf("martian: failed to write CONNECT response: %v", err)
			return err
		}
		brw.Flush()

		if cerr != nil {
			return cerr
		}

		defer dconn.Close()

		dbw := bufio.NewWriter(dconn)
		dbr := bufio.NewReader(dconn)
		defer dbw.Flush()

		copySync := func(w io.Writer, r io.Reader, donec chan<- bool) {
			io.Copy(w, r)
			donec <- true
		}

		donec := make(chan bool, 2)
		go copySync(dbw, brw, donec)
		go copySync(brw, dbr, donec)

		Debugf("martian: established CONNECT tunnel, proxying traffic")
		<-donec
		<-donec
		Debugf("martian: closed CONNECT tunnel")

		return closeConn
	}

	if err := p.reqmod.ModifyRequest(req); err != nil {
		Errorf("martian: error modifying request: %v", err)
		proxyutil.Warning(req.Header, err)
	}

	res, err := p.roundTrip(ctx, req)
	if err != nil {
		Errorf("martian: failed to round trip: %v", err)
		res = proxyutil.NewResponse(502, nil, req)
		proxyutil.Warning(res.Header, err)
	}

	if err := p.resmod.ModifyResponse(res); err != nil {
		Errorf("martian: error modifying response: %v", err)
		proxyutil.Warning(res.Header, err)
	}

	var closing error
	if req.Close || p.Closing() {
		Debugf("martian: received close request: %v", req.RemoteAddr)
		res.Header.Add("Connection", "close")
		closing = closeConn
	}

	Debugf("martian: sent response: %v", req.URL)
	if err := res.Write(brw); err != nil {
		Errorf("martian: failed to write response: %v", err)
		return err
	}
	brw.Flush()

	return closing
}

func (p *Proxy) roundTrip(ctx *session.Context, req *http.Request) (*http.Response, error) {
	if ctx.SkippingRoundTrip() {
		Debugf("martian: skipping round trip")
		return proxyutil.NewResponse(200, nil, req), nil
	}

	if tr, ok := p.roundTripper.(*http.Transport); ok {
		tr.Proxy = http.ProxyURL(p.proxyURL)
	}

	return p.roundTripper.RoundTrip(req)
}

func (p *Proxy) connect(req *http.Request) (*http.Response, net.Conn, error) {
	if p.proxyURL != nil {
		Debugf("martian: CONNECT with downstream proxy: %s", p.proxyURL.Host)

		conn, err := net.Dial("tcp", p.proxyURL.Host)
		if err != nil {
			return nil, nil, err
		}
		pbw := bufio.NewWriter(conn)
		pbr := bufio.NewReader(conn)

		if err := req.Write(pbw); err != nil {
			return nil, nil, err
		}
		pbw.Flush()

		res, err := http.ReadResponse(pbr, req)
		if err != nil {
			return nil, nil, err
		}

		return res, conn, nil
	}

	conn, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		return nil, nil, err
	}

	return proxyutil.NewResponse(200, nil, req), conn, nil
}
