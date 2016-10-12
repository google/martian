// Copyright 2016 Google Inc. All rights reserved.
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

// Package mobileproxy instantiates a Martian Proxy.
// This package is a reference implementation of Martian Proxy intended to
// be cross compiled with gomobile for use on Android and iOS.
package mobileproxy

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/google/martian"
	"github.com/google/martian/fifo"
	"github.com/google/martian/har"
	"github.com/google/martian/httpspec"
	mlog "github.com/google/martian/log"
	"github.com/google/martian/martianhttp"
	"github.com/google/martian/mitm"
	"github.com/google/martian/verify"

	// side-effect importing to register with JSON API
	_ "github.com/google/martian/body"
	_ "github.com/google/martian/cookie"
	_ "github.com/google/martian/fifo"
	_ "github.com/google/martian/header"
	_ "github.com/google/martian/martianurl"
	_ "github.com/google/martian/method"
	_ "github.com/google/martian/pingback"
	_ "github.com/google/martian/priority"
	_ "github.com/google/martian/querystring"
	_ "github.com/google/martian/skip"
	_ "github.com/google/martian/status"
)

// Martian is a wrapper for the initialized Martian proxy
type Martian struct {
	proxy    *martian.Proxy
	listener net.Listener
	mux      *http.ServeMux
}

// Start runs a martian.Proxy on addr
func Start(trafficPort, apiPort int) (*Martian, error) {
	return StartWithCertificate(trafficPort, apiPort, "", "")
}

// StartWithCertificate runs a proxy on addr and configures a cert for MITM
func StartWithCertificate(trafficPort int, apiPort int, cert, key string) (*Martian, error) {
	flag.Set("logtostderr", "true")

	p := martian.NewProxy()

	mux := http.NewServeMux()

	if cert != "" && key != "" {
		tlsc, err := tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			log.Fatal(err)
		}

		mlog.Debugf("mobileproxy: loaded cert and key")

		x509c, err := x509.ParseCertificate(tlsc.Certificate[0])
		if err != nil {
			log.Fatal(err)
		}

		mlog.Debugf("mobileproxy: parsed cert")

		mc, err := mitm.NewConfig(x509c, tlsc.PrivateKey)
		if err != nil {
			log.Fatal(err)
		}

		mc.SetValidity(12 * time.Hour)
		mc.SetOrganization("Martian Proxy")

		p.SetMITM(mc)

		handle(mux, "/authority.cer", apiPort, martianhttp.NewAuthorityHandler(x509c))
	}

	topg := fifo.NewGroup()

	stack, fg := httpspec.NewStack("martian.mobileproxy")
	topg.AddRequestModifier(stack)
	topg.AddResponseModifier(stack)

	p.SetRequestModifier(topg)
	p.SetResponseModifier(topg)

	// add HAR logger
	hl := har.NewLogger()
	stack.AddRequestModifier(hl)
	stack.AddResponseModifier(hl)

	m := martianhttp.NewModifier()
	fg.AddRequestModifier(m)
	fg.AddResponseModifier(m)

	// Proxy specific handlers.
	// These handlers take precendence over proxy traffic and will not be intercepted.

	// Retrieve HAR logs
	handle(mux, "/logs", apiPort, har.NewExportHandler(hl))
	handle(mux, "/logs/reset", apiPort, har.NewResetHandler(hl))

	// Update modifiers.
	handle(mux, "/configure", apiPort, m)

	// Verify assertions.
	vh := verify.NewHandler()
	vh.SetRequestVerifier(m)
	vh.SetResponseVerifier(m)
	handle(mux, "/verify", apiPort, vh)

	// Reset verifications.
	rh := verify.NewResetHandler()
	rh.SetRequestVerifier(m)
	rh.SetResponseVerifier(m)
	handle(mux, "/verify/reset", apiPort, rh)

	// Ignore SIGPIPE
	mlog.Debugf("mobileproxy: ignoring SIGPIPE signals")
	signal.Ignore(syscall.SIGPIPE)

	// start the API server
	apiAddr := fmt.Sprintf(":%d", apiPort)
	go http.ListenAndServe(apiAddr, nil)
	mlog.Infof("mobileproxy: proxy API started on %s", apiAddr)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", apiPort))
	if err != nil {
		return nil, err
	}

	mlog.Debugf("mobileproxy: started listener: %v", l.Addr())
	mlog.Infof("mobileproxy: starting proxy")
	go p.Serve(l)

	return &Martian{
		proxy:    p,
		listener: l,
		mux:      mux,
	}, nil
}

// Shutdown tells the Proxy to close. The proxy will stay alive until all connections through it
// have closed or timed out.
func (p *Martian) Shutdown() {
	mlog.Infof("mobileproxy: shutting down proxy")
	p.proxy.Close()
	mlog.Infof("mobileproxy: proxy shut down")
}

// SetLogLevel sets the Martian log level (Silent = 0, Error, Info, Debug), controlling which Martian
// log calls are displayed in the console
func SetLogLevel(l int) {
	mlog.SetLevel(l)
}

// handle sets up mux to handle requests to match patterns martian.proxy/{pth} and
// localhost:{apiPort}/{pth}. This assumes that the API server is running at
// localhost:{apiPort}, and requests to martian.proxy are forwarded there.
func handle(mux *http.ServeMux, pth string, apiPort int, handler http.Handler) {
	pattern := path.Join("martian.proxy", pth)
	mux.Handle(pattern, handler)
	mlog.Infof("mobileproxy: handler registered for %s", pattern)

	pattern = path.Join(fmt.Sprintf("localhost:%d", apiPort), pth)
	mux.Handle(pattern, handler)
	mlog.Infof("mobileproxy: handler registered for %s", pattern)
}
