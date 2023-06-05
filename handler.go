// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
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
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/proxyutil"
)

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func addTrailerHeader(rw http.ResponseWriter, tr http.Header) int {
	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	announcedTrailers := len(tr)
	if announcedTrailers == 0 {
		return 0
	}

	trailerKeys := make([]string, 0, announcedTrailers)
	for k := range tr {
		trailerKeys = append(trailerKeys, k)
	}
	rw.Header().Add("Trailer", strings.Join(trailerKeys, ", "))

	return announcedTrailers
}

func copyBody(w io.Writer, body io.ReadCloser) error {
	bufp := copyBufPool.Get().(*[]byte)
	buf := *bufp
	defer copyBufPool.Put(bufp)

	_, err := io.CopyBuffer(w, body, buf)
	return err
}

// proxyHandler wraps Proxy and implements http.Handler.
//
// Known limitations:
//   - MITM is not supported
//   - HTTP status code 100 is not supported, see [issue 2184]
//
// [issue 2184]: https://github.com/golang/go/issues/2184
type proxyHandler struct {
	*Proxy
}

// Handler returns proxy as http.Handler, see [proxyHandler] for details.
func (p *Proxy) Handler() http.Handler {
	return proxyHandler{p}
}

func (p proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	session := newSessionWithResponseWriter(rw)
	if req.TLS != nil {
		session.MarkSecure()
	}
	ctx := withSession(session)

	outreq := req.Clone(ctx.addToContext(req.Context()))
	if req.ContentLength == 0 {
		outreq.Body = http.NoBody
	}
	if outreq.Body != nil {
		defer outreq.Body.Close()
	}
	outreq.Close = false

	err := p.handleRequest(ctx, rw, outreq)
	if err != nil {
		res := p.errorResponse(req, err)
		defer res.Body.Close()
		writeResponse(rw, res)
	}
}

func (p proxyHandler) handleConnectRequest(ctx *Context, rw http.ResponseWriter, req *http.Request) error {
	session := ctx.Session()

	if err := p.reqmod.ModifyRequest(req); err != nil {
		log.Errorf("martian: error modifying CONNECT request: %v", err)
		p.warning(req.Header, err)
	}
	if session.Hijacked() {
		log.Debugf("martian: connection hijacked by request modifier")
		return nil
	}

	log.Debugf("martian: attempting to establish CONNECT tunnel: %s", req.URL.Host)
	var (
		res  *http.Response
		cr   io.Reader
		cw   io.WriteCloser
		cerr error
	)
	if p.ConnectPassthrough {
		pr, pw := io.Pipe()
		req.Body = pr
		defer req.Body.Close()

		// perform the HTTP roundtrip
		res, cerr = p.roundTrip(ctx, req)
		if res != nil {
			cr = res.Body
			cw = pw

			if res.StatusCode/100 == 2 {
				res = proxyutil.NewResponse(200, nil, req)
			}
		}
	} else {
		var cconn net.Conn
		res, cconn, cerr = p.connect(req)

		if cconn != nil {
			defer cconn.Close()
			cr = cconn
			cw = cconn
		}
	}

	if cerr != nil {
		log.Errorf("martian: failed to CONNECT: %v", cerr)
		res = p.errorResponse(req, cerr)
		p.warning(res.Header, cerr)
	}
	defer res.Body.Close()

	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Errorf("martian: error modifying CONNECT response: %v", err)
		p.warning(res.Header, err)
	}
	if session.Hijacked() {
		log.Debugf("martian: connection hijacked by response modifier")
		return nil
	}

	if res.StatusCode != 200 {
		if cerr == nil {
			log.Errorf("martian: CONNECT rejected with status code: %d", res.StatusCode)
		}
		writeResponse(rw, res)
		return nil
	}

	var (
		rc    = http.NewResponseController(rw)
		donec = make(chan bool, 2)
	)
	switch req.ProtoMajor {
	case 1:
		conn, brw, err := rc.Hijack()
		if err != nil {
			return err
		}
		defer conn.Close()

		if err := drainBuffer(cw, brw.Reader); err != nil {
			return err
		}

		res.ContentLength = -1
		if err := res.Write(brw); err != nil {
			log.Errorf("martian: got error while writing response back to client: %v", err)
		}
		if err := brw.Flush(); err != nil {
			log.Errorf("martian: got error while flushing response back to client: %v", err)
		}

		go copySync("outbound", cw, conn, donec)
		go copySync("inbound", conn, cr, donec)
	case 2:
		copyHeader(rw.Header(), res.Header)
		rw.WriteHeader(res.StatusCode)

		if err := rc.Flush(); err != nil {
			log.Errorf("martian: got error while flushing response back to client: %v", err)
		}

		go copySync("outbound", cw, req.Body, donec)
		go copySync("inbound", writeFlusher{rw, rc}, cr, donec)
	default:
		return fmt.Errorf("unsupported protocol version: %d", req.ProtoMajor)
	}

	log.Debugf("martian: established CONNECT tunnel, proxying traffic")
	<-donec
	<-donec
	log.Debugf("martian: closed CONNECT tunnel")

	return nil
}

// handleRequest handles a request and writes the response to the given http.ResponseWriter.
// It returns an error if the request
func (p proxyHandler) handleRequest(ctx *Context, rw http.ResponseWriter, req *http.Request) error {
	session := ctx.Session()

	if req.Method == "CONNECT" {
		return p.handleConnectRequest(ctx, rw, req)
	}

	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.RequestURI = ""

	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
		if session.IsSecure() {
			req.URL.Scheme = "https"
		}
	} else if req.URL.Scheme == "http" {
		if session.IsSecure() && !p.AllowHTTP {
			log.Infof("martian: forcing HTTPS inside secure session")
			req.URL.Scheme = "https"
		}
	}

	if err := p.reqmod.ModifyRequest(req); err != nil {
		log.Errorf("martian: error modifying request: %v", err)
		p.warning(req.Header, err)
	}
	if session.Hijacked() {
		log.Debugf("martian: connection hijacked by request modifier")
		return nil
	}

	// perform the HTTP roundtrip
	res, err := p.roundTrip(ctx, req)
	if err != nil {
		log.Errorf("martian: failed to round trip: %v", err)
		res = p.errorResponse(req, err)
		p.warning(res.Header, err)
	}
	defer res.Body.Close()

	// set request to original request manually, res.Request may be changed in transport.
	// see https://github.com/google/martian/issues/298
	res.Request = req

	if err := p.resmod.ModifyResponse(res); err != nil {
		log.Errorf("martian: error modifying response: %v", err)
		p.warning(res.Header, err)
	}
	if session.Hijacked() {
		log.Debugf("martian: connection hijacked by response modifier")
		return nil
	}

	if !req.ProtoAtLeast(1, 1) || req.Close || res.Close || p.Closing() {
		log.Debugf("martian: received close request: %v", req.RemoteAddr)
		res.Close = true
	}
	if p.CloseAfterReply {
		res.Close = true
	}

	writeResponse(rw, res)
	return nil
}

func newWriteFlusher(rw http.ResponseWriter) writeFlusher {
	return writeFlusher{
		rw: rw,
		rc: http.NewResponseController(rw),
	}
}

type writeFlusher struct {
	rw io.Writer
	rc *http.ResponseController
}

func (w writeFlusher) Write(p []byte) (n int, err error) {
	n, err = w.rw.Write(p)

	if n > 0 {
		if err := w.rc.Flush(); err != nil {
			log.Errorf("martian: got error while flushing response back to client: %v", err)
		}
	}

	return
}

func (w writeFlusher) CloseWrite() error {
	// This is a nop implementation of closeWriter.
	// It avoids printing the error log "cannot close write side of inbound CONNECT tunnel".
	return nil
}

func writeResponse(rw http.ResponseWriter, res *http.Response) {
	copyHeader(rw.Header(), res.Header)
	if res.Close {
		res.Header.Set("Connection", "close")
	}
	announcedTrailers := addTrailerHeader(rw, res.Trailer)
	rw.WriteHeader(res.StatusCode)

	// This flush is needed for http/1 server to flush the status code and headers.
	// It prevents the server from buffering the response and trying to calculate the response size.
	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	var err error
	if shouldFlush(res) {
		err = copyBody(newWriteFlusher(rw), res.Body)
	} else {
		err = copyBody(rw, res.Body)
	}
	if err != nil {
		log.Errorf("martian: got error while writing response back to client: %v", err)
		panic(http.ErrAbortHandler)
	}

	res.Body.Close() // close now, instead of defer, to populate res.Trailer
	if len(res.Trailer) == announcedTrailers {
		copyHeader(rw.Header(), res.Trailer)
	} else {
		h := rw.Header()
		for k, vv := range res.Trailer {
			for _, v := range vv {
				h.Add(http.TrailerPrefix+k, v)
			}
		}
	}
}
