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
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/google/martian"
	"github.com/google/martian/api"
	"github.com/google/martian/cybervillains"
	"github.com/google/martian/fifo"
	"github.com/google/martian/har"
	"github.com/google/martian/httpspec"
	mlog "github.com/google/martian/log"
	"github.com/google/martian/martianhttp"
	"github.com/google/martian/mitm"
	"github.com/google/martian/servemux"
	"github.com/google/martian/verify"

	// side-effect importing to register with JSON API
	_ "github.com/google/martian/body"
	_ "github.com/google/martian/cookie"
	_ "github.com/google/martian/header"
	_ "github.com/google/martian/martianurl"
	_ "github.com/google/martian/method"
	_ "github.com/google/martian/pingback"
	_ "github.com/google/martian/port"
	_ "github.com/google/martian/priority"
	_ "github.com/google/martian/querystring"
	_ "github.com/google/martian/skip"
	_ "github.com/google/martian/stash"
	_ "github.com/google/martian/static"
	_ "github.com/google/martian/status"
)

var started bool = false

// Martian is a wrapper for the initialized Martian proxy
type Martian struct {
	proxy    *martian.Proxy
	listener net.Listener
	mux      *http.ServeMux
}

// Start runs a martian.Proxy on trafficPort and the API server on apiPort.
func Start(trafficPort, apiPort int) (*Martian, error) {
	return StartWithCertificate(trafficPort, apiPort, "", "")
}

// StartWithCyberVillains runs a martian.Proxy on trafficPort and the API
// server on apiPort configured to perform MITM with the CyberVillains cert and key.
func StartWithCyberVillains(trafficPort int, apiPort int) (*Martian, error) {
	return StartWithCertificate(trafficPort, apiPort, cybervillains.Cert, cybervillains.Key)
}

// StartWithCertificate runs a martian.Proxy on trafficPort and the API
// server on apiPort configured to perform MITM with the cert and key provided.
func StartWithCertificate(trafficPort int, apiPort int, cert, key string) (*Martian, error) {
	var err error
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", trafficPort))
	if err != nil {
		log.Fatal(err)
	}

	mlog.Debugf("mobileproxy: started listener on: %v", listener.Addr())
	proxy := martian.NewProxy()
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

		proxy.SetMITM(mc)

		handle(mux, "/authority.cer", apiPort, martianhttp.NewAuthorityHandler(x509c))
	}

	// Forward traffic that pattern matches in http.DefaultServeMux before applying
	// httpspec modifiers (via modifier, specifically)
	topg := fifo.NewGroup()
	apif := servemux.NewFilter(nil)
	apif.SetRequestModifier(api.NewForwarder("", apiPort))
	topg.AddRequestModifier(apif)

	stack, fg := httpspec.NewStack("martian.mobileproxy")
	topg.AddRequestModifier(stack)
	topg.AddResponseModifier(stack)

	proxy.SetRequestModifier(topg)
	proxy.SetResponseModifier(topg)

	// add HAR logger for unmodified logs.
	uhl := har.NewLogger()
	fg.AddRequestModifier(uhl)
	fg.AddResponseModifier(uhl)

	// add HAR logger
	hl := har.NewLogger()
	stack.AddRequestModifier(hl)
	stack.AddResponseModifier(hl)

	m := martianhttp.NewModifier()
	fg.AddRequestModifier(m)
	fg.AddResponseModifier(m)

	// Proxy specific handlers.
	// These handlers take precendence over proxy traffic and will not be intercepted.

	// Update modifiers.
	handle(mux, "/configure", apiPort, m)
	mlog.Infof("mobileproxy: configure with requests to http://martian.proxy/configure")

	// Retrieve Unmodified HAR logs
	handle(mux, "/logs/original", apiPort, har.NewExportHandler(uhl))
	handle(mux, "/logs/original/reset", apiPort, har.NewResetHandler(uhl))

	// Retrieve HAR logs
	handle(mux, "/logs", apiPort, har.NewExportHandler(hl))
	handle(mux, "/logs/reset", apiPort, har.NewResetHandler(hl))

	// Verify assertions.
	vh := verify.NewHandler()
	vh.SetRequestVerifier(m)
	vh.SetResponseVerifier(m)

	handle(mux, "/verify", apiPort, vh)
	mlog.Infof("mobileproxy: check verifications with requests to http://martian.proxy/verify")

	// Reset verifications.
	rh := verify.NewResetHandler()
	rh.SetRequestVerifier(m)
	rh.SetResponseVerifier(m)
	handle(mux, "/verify/reset", apiPort, rh)
	mlog.Infof("mobileproxy: reset verifications with requests to http://martian.proxy/verify/reset")

	mlog.Infof("mobileproxy: starting Martian proxy on listener")
	go proxy.Serve(listener)

	// start the API server
	apiAddr := fmt.Sprintf(":%d", apiPort)
	go http.ListenAndServe(apiAddr, mux)
	mlog.Infof("mobileproxy: proxy API started on %s", apiAddr)
	started = true

	return &Martian{
		proxy:    proxy,
		listener: listener,
		mux:      mux,
	}, nil
}

func IsStarted() bool {
	return started
}

// Shutdown tells the Proxy to close. The proxy will stay alive until all connections through it
// have closed or timed out.
func (p *Martian) Shutdown() {
	mlog.Infof("mobileproxy: shutting down proxy")
	p.listener.Close()
	p.proxy.Close()
	started = false
	mlog.Infof("mobileproxy: proxy shut down")
}

// SetLogLevel sets the Martian log level (Silent = 0, Error, Info, Debug), controlling which Martian
// log calls are displayed in the console
func SetLogLevel(l int) {
	mlog.SetLevel(l)
}

// handle sets up http.DefaultServeMux to handle requests to match patterns martian.proxy/{pth} and
// localhost:{apiPort}/{pth}. This assumes that the API server is running at localhost:{apiPort}, and
// requests to martian.proxy are forwarded there.
func handle(mux *http.ServeMux, pattern string, apiPort int, handler http.Handler) {
	mux.Handle(pattern, handler)
	mlog.Infof("mobileproxy: handler registered for %s", pattern)

	lhp := path.Join(fmt.Sprintf("localhost:%d", apiPort), pattern)
	mux.Handle(lhp, handler)
	mlog.Infof("mobileproxy: handler registered for %s", lhp)
}
