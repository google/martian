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
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/martian/martiantest"
	"github.com/google/martian/mitm"
	"github.com/google/martian/session"
)

type tempError struct{}

func (e *tempError) Error() string   { return "temporary" }
func (e *tempError) Timeout() bool   { return true }
func (e *tempError) Temporary() bool { return true }

type timeoutListener struct {
	net.Listener
	errCount int
	err      error
}

func newTimeoutListener(l net.Listener, errCount int) net.Listener {
	return &timeoutListener{
		Listener: l,
		errCount: errCount,
		err:      &tempError{},
	}
}

func (l *timeoutListener) Accept() (net.Conn, error) {
	if l.errCount > 0 {
		l.errCount--
		return nil, l.err
	}

	return l.Listener.Accept()
}

func TestContext(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	want := session.NewContext()
	SetContext(req, want)

	if got := Context(req); got != want {
		t.Errorf("Context(req): got %v, want %v", got, want)
	}

	RemoveContext(req)

	if got := Context(req); got != nil {
		t.Errorf("Context(req): got %v, want nil", got)
	}
}

func TestIntegrationTemporaryTimeout(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Liste(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	tr := martiantest.NewTransport()
	p.Option(RoundTripper(tr))
	p.Option(Timeout(100 * time.Millisecond))

	// Start the proxy with a listener that will return a temporary error on
	// Accept() three times.
	go p.Serve(newTimeoutListener(l, 3))

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	req.Header.Set("Connection", "close")

	// GET http://example.com/ HTTP/1.1
	// Host: example.com
	if err := req.WriteProxy(conn); err != nil {
		t.Fatalf("req.WriteProxy(): got %v, want no error", err)
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
}

func TestIntegrationHTTP(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	tr := martiantest.NewTransport()
	p.Option(RoundTripper(tr))

	tm := martiantest.NewModifier()

	tm.RequestFunc(func(req *http.Request) {
		ctx := Context(req)
		ctx.Set("martian.test", "true")
	})

	tm.ResponseFunc(func(res *http.Response) {
		ctx := Context(res.Request)
		v, _ := ctx.Get("martian.test")

		res.Header.Set("Martian-Test", v.(string))
	})

	p.SetRequestModifier(tm)
	p.SetResponseModifier(tm)

	go p.Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// GET http://example.com/ HTTP/1.1
	// Host: example.com
	if err := req.WriteProxy(conn); err != nil {
		t.Fatalf("req.WriteProxy(): got %v, want no error", err)
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	if got, want := res.StatusCode, 200; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}

	if got, want := res.Header.Get("Martian-Test"), "true"; got != want {
		t.Errorf("res.Header.Get(%q): got %q, want %q", "Martian-Test", got, want)
	}
}

func TestIntegrationHTTPDownstreamProxy(t *testing.T) {
	// Start first proxy to use as downstream.
	dl, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	downstream := NewProxy()
	defer downstream.Close()

	dtr := martiantest.NewTransport()
	dtr.Respond(299)
	downstream.Option(RoundTripper(dtr))
	downstream.Option(Timeout(50 * time.Millisecond))

	go downstream.Serve(dl)

	// Start second proxy as upstream proxy, will write to downstream proxy.
	ul, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	upstream := NewProxy()
	defer upstream.Close()

	// Set upstream proxy's downstream proxy to the host:port of the first proxy.
	upstream.Option(DownstreamProxy(&url.URL{
		Host: dl.Addr().String(),
	}))
	upstream.Option(Timeout(50 * time.Millisecond))

	go upstream.Serve(ul)

	// Open connection to upstream proxy.
	conn, err := net.Dial("tcp", ul.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// GET http://example.com/ HTTP/1.1
	// Host: example.com
	if err := req.WriteProxy(conn); err != nil {
		t.Fatalf("req.WriteProxy(): got %v, want no error", err)
	}

	// Response from downstream proxy.
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	if got, want := res.StatusCode, 299; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}
}

func TestIntegrationHTTPDownstreamProxyError(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	// Set proxy's downstream proxy to invalid host:port to force failure.
	p.Option(DownstreamProxy(&url.URL{
		Host: "[::1]:0",
	}))
	p.Option(Timeout(50 * time.Millisecond))

	go p.Serve(l)

	// Open connection to upstream proxy.
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("CONNECT", "//www.google.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// CONNECT www.google.com:443 HTTP/1.1
	// Host: www.google.com
	if err := req.Write(conn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	// Response from upstream proxy, assuming downstream proxy failed to CONNECT.
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	if got, want := res.StatusCode, 502; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}
}

func TestIntegrationConnect(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	tm := martiantest.NewModifier()
	reqerr := errors.New("request error")
	reserr := errors.New("response error")

	tm.RequestError(reqerr)
	tm.ResponseError(reserr)

	p.SetRequestModifier(tm)
	p.SetResponseModifier(tm)

	go p.Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("CONNECT", "//www.google.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// CONNECT www.google.com:443 HTTP/1.1
	// Host: www.google.com:443
	if err := req.Write(conn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	// CONNECT response after establishing tunnel to production.
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	if got, want := res.StatusCode, 200; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}

	if !tm.RequestModified() {
		t.Error("tm.RequestModified(): got false, want true")
	}
	if !tm.ResponseModified() {
		t.Error("tm.ResponseModified(): got false, want true")
	}
	if got, want := res.Header.Get("Warning"), reserr.Error(); !strings.Contains(got, want) {
		t.Errorf("res.Header.Get(%q): got %s, want to contain %s", got, want)
	}

	tlsconn := tls.Client(conn, &tls.Config{
		// Validate the certificate is actually for Google.
		ServerName: "www.google.com",
	})
	defer tlsconn.Close()

	req, err = http.NewRequest("GET", "https://www.google.com/humans.txt", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// Request is tunneled to a real Google server.
	// TODO: Avoid creating real connection, spin up TLS server with fake
	// certificate signed by generated CA.
	//
	// GET /humans.text HTTP/1.1
	// Host: www.google.com
	if err := req.Write(tlsconn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err = http.ReadResponse(bufio.NewReader(tlsconn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}
	if got, want := res.Header.Get("Warning"), reserr.Error(); strings.Contains(got, want) {
		t.Errorf("res.Header.Get(%q): got %s, want to not contain %s", got, want)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}

	want := []byte("Google is built by a large team of engineers")

	if !bytes.Contains(got, want) {
		t.Fatalf("res.Body: got %q, want to contain %q", got, want)
	}
}

func TestIntegrationConnectDownstreamProxy(t *testing.T) {
	// Start first proxy to use as downstream.
	dl, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	downstream := NewProxy()
	defer downstream.Close()

	dtr := martiantest.NewTransport()
	dtr.Respond(299)
	downstream.Option(RoundTripper(dtr))
	downstream.Option(Timeout(50 * time.Millisecond))

	ca, priv, err := mitm.NewAuthority("martian.proxy", "Martian Authority", 2*time.Hour)
	if err != nil {
		t.Fatalf("mitm.NewAuthority(): got %v, want no error", err)
	}

	mc, err := mitm.NewConfig(ca, priv)
	if err != nil {
		t.Fatalf("mitm.NewConfig(): got %v, want no error", err)
	}
	downstream.Option(MITM(mc))

	go downstream.Serve(dl)

	// Start second proxy as upstream proxy, will CONNECT to downstream proxy.
	ul, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	upstream := NewProxy()
	defer upstream.Close()

	// Set upstream proxy's downstream proxy to the host:port of the first proxy.
	upstream.Option(DownstreamProxy(&url.URL{
		Host: dl.Addr().String(),
	}))
	upstream.Option(Timeout(50 * time.Millisecond))

	go upstream.Serve(ul)

	// Open connection to upstream proxy.
	conn, err := net.Dial("tcp", ul.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("CONNECT", "//www.google.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// CONNECT www.google.com:443 HTTP/1.1
	// Host: www.google.com
	if err := req.Write(conn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	// Response from downstream proxy starting MITM.
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	if got, want := res.StatusCode, 200; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}

	roots := x509.NewCertPool()
	roots.AddCert(ca)

	tlsconn := tls.Client(conn, &tls.Config{
		// Validate the hostname is Google.
		ServerName: "www.google.com",
		// The certificate will have been MITM'd, verify using the MITM CA
		// certificate.
		RootCAs: roots,
	})
	defer tlsconn.Close()

	req, err = http.NewRequest("GET", "https://www.google.com/humans.txt", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// GET /humans.txt HTTP/1.1
	// Host: www.google.com
	if err := req.Write(tlsconn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	// Response from MITM in downstream proxy.
	res, err = http.ReadResponse(bufio.NewReader(tlsconn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 299; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}
}

func TestIntegrationMITM(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	tr := martiantest.NewTransport()
	p.Option(RoundTripper(tr))

	ca, priv, err := mitm.NewAuthority("martian.proxy", "Martian Authority", 2*time.Hour)
	if err != nil {
		t.Fatalf("mitm.NewAuthority(): got %v, want no error", err)
	}

	mc, err := mitm.NewConfig(ca, priv)
	if err != nil {
		t.Fatalf("mitm.NewConfig(): got %v, want no error", err)
	}
	p.Option(MITM(mc))

	tm := martiantest.NewModifier()
	reqerr := errors.New("request error")
	reserr := errors.New("response error")
	tm.RequestError(reqerr)
	tm.ResponseError(reserr)

	p.SetRequestModifier(tm)
	p.SetResponseModifier(tm)

	go p.Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("CONNECT", "//www.google.com:443", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// CONNECT www.google.com:443 HTTP/1.1
	// Host: www.google.com
	if err := req.Write(conn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	// Response MITM'd from proxy.
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	if got, want := res.StatusCode, 200; got != want {

		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	if got, want := res.Header.Get("Warning"), reserr.Error(); !strings.Contains(got, want) {
		t.Errorf("res.Header.Get(%q): got %s, want to contain %s", "Warning", got, want)
	}

	roots := x509.NewCertPool()
	roots.AddCert(ca)

	tlsconn := tls.Client(conn, &tls.Config{
		ServerName: "www.google.com",
		RootCAs:    roots,
	})
	defer tlsconn.Close()

	req, err = http.NewRequest("GET", "https://www.google.com/humans.txt", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// GET /humans.text HTTP/1.1
	// Host: www.google.com
	if err := req.Write(tlsconn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	// Response from MITM proxy.
	res, err = http.ReadResponse(bufio.NewReader(tlsconn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
	if got, want := res.Header.Get("Warning"), reserr.Error(); !strings.Contains(got, want) {
		t.Errorf("res.Header.Get(%q): got %s, want to contain %s", "Warning", got, want)
	}
}

func TestIntegrationTransparentHTTP(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	tr := martiantest.NewTransport()
	p.Option(RoundTripper(tr))

	tm := martiantest.NewModifier()
	p.SetRequestModifier(tm)
	p.SetResponseModifier(tm)

	go p.Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// GET / HTTP/1.1
	// Host: www.example.com
	if err := req.Write(conn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}

	if got, want := res.StatusCode, 200; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}

	if !tm.RequestModified() {
		t.Error("tm.RequestModified(): got false, want true")
	}
	if !tm.ResponseModified() {
		t.Error("tm.ResponseModified(): got false, want true")
	}
}

func TestIntegrationTransparentMITM(t *testing.T) {
	ca, priv, err := mitm.NewAuthority("martian.proxy", "Martian Authority", 2*time.Hour)
	if err != nil {
		t.Fatalf("mitm.NewAuthority(): got %v, want no error", err)
	}

	mc, err := mitm.NewConfig(ca, priv)
	if err != nil {
		t.Fatalf("mitm.NewConfig(): got %v, want no error", err)
	}

	// Start TLS listener with config that will generate certificates based on
	// SNI from connection.
	l, err := tls.Listen("tcp", "[::1]:0", mc.TLS())
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	tr := martiantest.NewTransport()
	p.Option(RoundTripper(tr))

	tm := martiantest.NewModifier()
	p.SetRequestModifier(tm)
	p.SetResponseModifier(tm)

	go p.Serve(l)

	roots := x509.NewCertPool()
	roots.AddCert(ca)

	tlsconn, err := tls.Dial("tcp", l.Addr().String(), &tls.Config{
		// Verify the hostname is www.google.com.
		ServerName: "www.google.com",
		// The certificate will have been generated during MITM, so we need to
		// verify it with the generated CA certificate.
		RootCAs: roots,
	})
	if err != nil {
		t.Fatalf("tls.Dial(): got %v, want no error", err)
	}
	defer tlsconn.Close()

	req, err := http.NewRequest("GET", "https://www.google.com/humans.txt", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// Write Encrypted request directly, no CONNECT.
	// GET /humans.txt HTTP/1.1
	// Host: www.google.com
	if err := req.Write(tlsconn); err != nil {
		t.Fatalf("req.Write(): got %v, want no error", err)
	}

	res, err := http.ReadResponse(bufio.NewReader(tlsconn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Fatalf("res.StatusCode: got %d, want %d", got, want)
	}

	if !tm.RequestModified() {
		t.Errorf("tm.RequestModified(): got false, want true")
	}
	if !tm.ResponseModified() {
		t.Errorf("tm.ResponseModified(): got false, want true")
	}
}

func TestIntegrationFailedRoundTrip(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	tr := martiantest.NewTransport()
	trerr := errors.New("round trip error")
	tr.RespondError(trerr)
	p.Option(RoundTripper(tr))

	go p.Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// GET http://example.com/ HTTP/1.1
	// Host: example.com
	if err := req.WriteProxy(conn); err != nil {
		t.Fatalf("req.WriteProxy(): got %v, want no error", err)
	}

	// Response from failed round trip.
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 502; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}

	if got, want := res.Header.Get("Warning"), trerr.Error(); !strings.Contains(got, want) {
		t.Errorf("res.Header.Get(%q): got %s, want to contain %s", "Warning", got, want)
	}
}

func TestIntegrationSkipRoundTrip(t *testing.T) {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	p := NewProxy()
	defer p.Close()

	// Transport will be skipped, no 500.
	tr := martiantest.NewTransport()
	tr.Respond(500)
	p.Option(RoundTripper(tr))

	tm := martiantest.NewModifier()
	tm.RequestFunc(func(req *http.Request) {
		ctx := Context(req)
		ctx.SkipRoundTrip()
	})
	p.SetRequestModifier(tm)

	go p.Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	// GET http://example.com/ HTTP/1.1
	// Host: example.com
	if err := req.WriteProxy(conn); err != nil {
		t.Fatalf("req.WriteProxy(): got %v, want no error", err)
	}

	// Response from skipped round trip.
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("http.ReadResponse(): got %v, want no error", err)
	}
	defer res.Body.Close()

	if got, want := res.StatusCode, 200; got != want {
		t.Errorf("res.StatusCode: got %d, want %d", got, want)
	}
}
