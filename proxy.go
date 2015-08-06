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

var errClose = errors.New("closing connection")
var noop = Noop("martian")

var (
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

func isCloseable(err error) bool {
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return true
	}

	switch err {
	case io.EOF, io.ErrClosedPipe, errClose:
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
	mux          *http.ServeMux
	closing      int32 // atomic

	reqmod RequestModifier
	resmod ResponseModifier
}

// NewProxy returns a new HTTP proxy.
func NewProxy() *Proxy {
	return &Proxy{
		roundTripper: http.DefaultTransport,
		timeout:      5 * time.Minute,
		conns:        &sync.WaitGroup{},
		mux:          http.DefaultServeMux,
		reqmod:       noop,
		resmod:       noop,
	}
}

// SetRoundTripper sets the http.RoundTripper of the proxy.
func (p *Proxy) SetRoundTripper(rt http.RoundTripper) {
	p.roundTripper = rt
}

// SetTimeout sets the request timeout of the proxy.
func (p *Proxy) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}

// SetMITM sets the config to use for MITMing of CONNECT requests.
func (p *Proxy) SetMITM(config *mitm.Config) {
	p.mitm = config
}

// SetDownstreamProxy sets the proxy that receives requests from the upstream
// proxy.
func (p *Proxy) SetDownstreamProxy(proxyURL *url.URL) {
	p.proxyURL = proxyURL
}

// SetMux sets the http.ServeMux to use for internal proxy requests. Defaults
// to http.DefaultServeMux.
func (p *Proxy) SetMux(mux *http.ServeMux) {
	p.mux = mux
}

// Close sets the proxying to the closing state and waits for all connections
// to resolve.
func (p *Proxy) Close() {
	Infof("martian: closing down proxy")

	atomic.StoreInt32(&p.closing, 1)
	p.conns.Wait()
}

// Closing returns whether the proxy is in the closing state.
func (p *Proxy) Closing() bool {
	return atomic.LoadInt32(&p.closing) == 1
}

// SetRequestModifier sets the request modifier.
func (p *Proxy) SetRequestModifier(reqmod RequestModifier) {
	if reqmod == nil {
		reqmod = noop
	}

	p.reqmod = reqmod
}

// SetResponseModifier sets the response modifier.
func (p *Proxy) SetResponseModifier(resmod ResponseModifier) {
	if resmod == nil {
		resmod = noop
	}

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

	ctx := session.FromContext(nil)
	brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		deadline := time.Now().Add(p.timeout)
		conn.SetDeadline(deadline)

		if err := p.handle(ctx, conn, brw); isCloseable(err) {
			Infof("martian: closing connection: %v", conn.RemoteAddr())
			return
		}
	}
}

func (p *Proxy) handle(ctx *session.Context, conn net.Conn, brw *bufio.ReadWriter) error {
	Debugf("martian: waiting for request: %v", conn.RemoteAddr())

	req, err := http.ReadRequest(brw.Reader)
	if err != nil {
		Errorf("martian: failed to read request: %v", err)

		// TODO: TCPConn.WriteClose() to avoid sending an RST to the client.

		return errClose
	}
	defer req.Body.Close()

	if h, pattern := p.mux.Handler(req); pattern != "" {
		defer brw.Flush()

		Infof("martian: received proxy specific request: %s", req.URL)
		rw := newResponseWriter(brw)
		defer rw.Close()

		h.ServeHTTP(rw, req)

		// Call WriteHeader to ensure a response is sent, since the handler isn't
		// required to call WriteHeader/Write.
		rw.WriteHeader(200)

		if req.Close {
			return errClose
		}

		return nil
	}

	ctx = session.FromContext(ctx)
	SetContext(req, ctx)
	defer RemoveContext(req)

	if tconn, ok := conn.(*tls.Conn); ok {
		ctx.GetSession().MarkSecure()

		cs := tconn.ConnectionState()
		req.TLS = &cs
	}

	req.URL.Scheme = "http"
	if ctx.GetSession().IsSecure() {
		Debugf("martian: forcing HTTPS inside secure session")
		req.URL.Scheme = "https"
	}

	req.RemoteAddr = conn.RemoteAddr().String()
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	Debugf("martian: received request: %s", req.URL)

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

			res.Write(brw)
			brw.Flush()

			Debugf("martian: completed MITM for connection: %s", req.Host)

			tlsconn := tls.Server(conn, p.mitm.TLSForHost(req.Host))
			brw.Writer.Reset(tlsconn)
			brw.Reader.Reset(tlsconn)

			return p.handle(ctx, tlsconn, brw)
		}

		Debugf("martian: attempting to establish CONNECT tunnel: %s", req.URL.Host)
		res, cconn, cerr := p.connect(req)
		if cerr != nil {
			Errorf("martian: failed to CONNECT: %v", err)
			res = proxyutil.NewResponse(502, nil, req)
			proxyutil.Warning(res.Header, cerr)

			if err := p.resmod.ModifyResponse(res); err != nil {
				Errorf("martian: error modifying CONNECT response: %v", err)
				proxyutil.Warning(res.Header, err)
			}

			res.Write(brw)
			return brw.Flush()
		}
		defer res.Body.Close()
		defer cconn.Close()

		if err := p.resmod.ModifyResponse(res); err != nil {
			Errorf("martian: error modifying CONNECT response: %v", err)
			proxyutil.Warning(res.Header, err)
		}

		res.Write(brw)
		brw.Flush()

		cbw := bufio.NewWriter(cconn)
		cbr := bufio.NewReader(cconn)
		defer cbw.Flush()

		copySync := func(w io.Writer, r io.Reader, donec chan<- bool) {
			io.Copy(w, r)
			donec <- true
		}

		donec := make(chan bool, 2)
		go copySync(cbw, brw, donec)
		go copySync(brw, cbr, donec)

		Debugf("martian: established CONNECT tunnel, proxying traffic")
		<-donec
		<-donec
		Debugf("martian: closed CONNECT tunnel")

		return errClose
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
	defer res.Body.Close()

	if err := p.resmod.ModifyResponse(res); err != nil {
		Errorf("martian: error modifying response: %v", err)
		proxyutil.Warning(res.Header, err)
	}

	var closing error
	if req.Close || p.Closing() {
		Debugf("martian: received close request: %v", req.RemoteAddr)
		res.Header.Add("Connection", "close")
		closing = errClose
	}

	Debugf("martian: sent response: %v", req.URL)
	res.Write(brw)
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

		req.Write(pbw)
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
