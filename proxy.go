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
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/martian/proxyutil"
)

// Proxy implements an HTTP proxy with CONNECT and TLS MITM support.
type Proxy struct {
	// RoundTripper used to make the request from the proxy to the target server.
	RoundTripper http.RoundTripper
	// Timeout is the length of time the connection will be kept open while idle.
	Timeout time.Duration

	mitm    *MITM
	creqmod RequestModifier
	cresmod ResponseModifier
	reqmod  RequestModifier
	resmod  ResponseModifier
}

// NewProxy returns an HTTP proxy.
func NewProxy(mitm *MITM) *Proxy {
	return &Proxy{
		RoundTripper: http.DefaultTransport,
		Timeout:      5 * time.Minute,
		mitm:         mitm,
	}
}

// SetConnectRequestModifier sets the request modifier for the CONNECT request.
func (p *Proxy) SetConnectRequestModifier(creqmod RequestModifier) {
	Debugf("set CONNECT request modifier")
	p.creqmod = creqmod
}

// SetConnectResponseModifier sets the response modifier for the CONNECT response.
func (p *Proxy) SetConnectResponseModifier(cresmod ResponseModifier) {
	Debugf("set CONNECT response modifier")
	p.cresmod = cresmod
}

// SetRequestModifier sets the request modifier for the decrypted request.
func (p *Proxy) SetRequestModifier(reqmod RequestModifier) {
	Debugf("set request modifier")
	p.reqmod = reqmod
}

// SetResponseModifier sets the response modifier for the decrypted response.
func (p *Proxy) SetResponseModifier(resmod ResponseModifier) {
	Debugf("set response modifier")
	p.resmod = resmod
}

// ModifyRequest modifies the request.
func (p *Proxy) ModifyRequest(ctx *Context, req *http.Request) error {
	if p.reqmod == nil {
		Debugf("no modifier set, skip modifying request %s", req.URL)
		return nil
	}

	Debugf("modifying request %s", req.URL)

	return p.reqmod.ModifyRequest(ctx, req)
}

// ModifyResponse modifies the response.
func (p *Proxy) ModifyResponse(ctx *Context, res *http.Response) error {
	if p.resmod == nil {
		Debugf("no modifier set, skip modifying response for %s", res.Request.URL)
		return nil
	}

	Debugf("modifying response %s", res.Request.URL)

	return p.resmod.ModifyResponse(ctx, res)
}

// ServeHTTP handles requests from a connection and writes responses.
//
// If a MITM config was provided and a CONNECT request is received, the proxy
// will generate a fake TLS certificate using the given authority certificate
// and perform the TLS handshake. The request will then be decrypted and
// modifiers will be run, followed by the request being re-encrypted and sent
// to the destination server.
//
// If no MITM config was provided and a CONNECT request is received, the proxy
// will open a connection to the destination server and copy the encrypted bytes
// directly, as per normal CONNECT semantics.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext()

	hj, ok := w.(http.Hijacker)
	if !ok {
		Errorf("w.(http.Hijacker): !ok")
		http.Error(w, "error unsupported http.Hijacker", 500)
		return
	}

	// Take over the connection immediately. We technically don't need to do this
	// in a non-CONNECT request, but it's easier to have all cases share the same
	// logic for request handling.
	conn, rw, err := hj.Hijack()
	if err != nil {
		Errorf("hj.Hijack(): %v", err)
		return
	}
	defer conn.Close()

	var closing bool
	switch r.Method {
	case "CONNECT":
		if r.URL.Host == "" {
			r.URL.Host = r.Host
		}

		// Run the CONNECT modifiers and handle errors.
		res, err := p.connectResponse(ctx, r)
		if err != nil {
			Errorf("connectResponse(%s): %v", r.URL, err)
			res.Write(rw)
			break
		}

		var tlsconn net.Conn
		var tlsrw *bufio.ReadWriter
		if p.mitm != nil {
			// Drop the port when building the MITM certificate.
			host, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				Errorf("net.SplitHostPort(%s): %v", r.Host, err)
				proxyutil.NewErrorResponse(400, err, r).Write(rw)
				break
			}

			// Build MITM certificate and wrap connection.
			tlsconn, tlsrw, err = p.mitm.Hijack(conn, host)
			if err != nil {
				Errorf("mitm.Hijack(conn, %s): %v", host, err)
				proxyutil.NewErrorResponse(400, err, r).Write(rw)
				break
			}
			defer tlsconn.Close()

			Debugf("Hijacked TLS connection for %s", host)
		}

		res.Write(rw)
		rw.Flush()

		if tlsconn != nil && tlsrw != nil {
			conn = tlsconn
			rw = tlsrw
		}

		// Proxy is not configured for man-in-the-middle and we have received a
		// CONNECT request. Proxy the request as normal CONNECT.
		if p.mitm == nil {
			p.handleNonMITMConnect(ctx, conn, r.Host)
			return
		}
	default:
		Debugf("received non-CONNECT request %s", r.URL)
		// Run the request and response modifiers.
		closing = p.handleRequest(ctx, rw, r)
	}

	Debugf("rw.Flush(): flushing response: %s", r.URL)
	if err := rw.Flush(); err != nil {
		Errorf("rw.Flush(): %v", err)
	}

	if closing {
		Debugf("closing connection")
		return
	}

	// We continue looping until the connection has been closed by the client.
	for {
		// Each request has its own timeout of p.Timeout. Reset after each request.
		deadline := time.Now().Add(p.Timeout)

		if err := conn.SetDeadline(deadline); err != nil {
			Errorf("conn.SetDeadline(%s): %v", deadline.Format(time.RFC3339), err)
			return
		}

		Debugf("Waiting for request...")
		req, err := http.ReadRequest(rw.Reader)
		if err != nil {
			// We have encountered a timeout error, do not attempt to send an error
			// response, just close the connection.
			neterr, ok := err.(net.Error)
			switch {
			case ok && neterr.Timeout():
			case err == io.EOF:
			case err == io.ErrClosedPipe:
			default:
				Errorf("http.ReadRequest(): %v", err)
				return
			}

			Debugf("http.ReadRequest(): timeout error %v", err)
			return
		}
		// Scheme will be empty in the case of a CONNECT request,
		// default to HTTPS if we don't have an original scheme.
		req.URL.Scheme = r.URL.Scheme
		if req.URL.Scheme == "" {
			req.URL.Scheme = "https"
		}

		// For requests received during MITM the URL.Host will not be set as it
		// does not appear in the Request-URI line. The http package will fill the
		// Host field with either the value from URL.Host or the Host header.
		req.URL.Host = req.Host

		req.RemoteAddr = conn.RemoteAddr().String()

		// Run the request and response modifiers.
		closing = p.handleRequest(ctx, rw, req)

		if err := rw.Flush(); err != nil {
			Errorf("rw.Flush(): %v", err)
		}

		if closing {
			Debugf("closing connection")
			return
		}
	}
}

// shouldCloseAfterReply returns whether the connection should be closed after
// the response has been sent.
func shouldCloseAfterReply(header http.Header) bool {
	for _, vs := range header["Connection"] {
		for _, v := range strings.Split(vs, ",") {
			if strings.ToLower(strings.TrimSpace(v)) == "close" {
				return true
			}
		}
	}

	return false
}

// connectResponse builds the CONNECT response and runs the CONNECT request and
// response modifiers.
func (p *Proxy) connectResponse(ctx *Context, req *http.Request) (*http.Response, error) {
	res := proxyutil.NewResponse(200, nil, req)

	if p.creqmod != nil {
		if err := p.creqmod.ModifyRequest(ctx, req); err != nil {
			Errorf("ModifyConnectRequest(%s): %v", req.URL, err)
			return proxyutil.NewErrorResponse(400, err, req), err
		}
	}

	if p.cresmod != nil {
		if err := p.cresmod.ModifyResponse(ctx, res); err != nil {
			Errorf("ModifyConnectResponse(%s): %v", res.Request.URL, err)
			return proxyutil.NewErrorResponse(400, err, req), err
		}
	}

	return res, nil
}

// handleNonMITMConnect dials the destination server and passes through the
// encrypted data from the client.
func (p *Proxy) handleNonMITMConnect(ctx *Context, conn net.Conn, host string) {
	Debugf("no MITM config found, fallback to normal CONNECT flow for %s", host)

	dc, err := net.Dial("tcp", host)
	if err != nil {
		Errorf("net.Dial(%q, %s, nil): %v", "tcp", host, err)
		return
	}
	defer dc.Close()

	Debugf("begin copy for %s", host)

	donec := make(chan bool, 1)

	go func() {
		io.Copy(dc, conn)
		donec <- true
	}()

	go func() {
		io.Copy(conn, dc)
		donec <- true
	}()

	<-donec

	Debugf("end copy for %s", host)
}

// handleRequest runs the request and response modifiers and performs the roundtrip to the destination server.
func (p *Proxy) handleRequest(ctx *Context, rw *bufio.ReadWriter, req *http.Request) (closing bool) {
	if shouldCloseAfterReply(req.Header) {
		Debugf("closing after reply")
		closing = true
	}

	if err := p.ModifyRequest(ctx, req); err != nil {
		Errorf("martian.ModifyRequest(): %v", err)
		proxyutil.NewErrorResponse(400, err, req).Write(rw)
		return
	}

	var res *http.Response
	var err error
	if !ctx.SkipRoundTrip {
		Debugf("proceed to round trip for %s", req.URL)

		res, err = p.RoundTripper.RoundTrip(req)
		if err != nil {
			Errorf("RoundTripper.RoundTrip(%s): %v", req.URL, err)
			proxyutil.NewErrorResponse(502, err, req).Write(rw)
			return
		}
	} else {
		Debugf("skipped round trip for %s", req.URL)
		res = proxyutil.NewResponse(200, nil, req)
	}

	if err := p.ModifyResponse(ctx, res); err != nil {
		Errorf("martian.ModifyResponse(): %v", err)
		proxyutil.NewErrorResponse(400, err, req).Write(rw)
		return
	}

	if closing {
		res.Header.Set("Connection", "close")
		res.Close = true
	}

	if err := res.Write(rw); err != nil {
		Errorf("res.Write(): %v", err)
	}

	return
}
