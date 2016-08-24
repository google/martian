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
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/martian"
	// side-effect importing to register with JSON API
	_ "github.com/google/martian/body"
	_ "github.com/google/martian/cookie"
	_ "github.com/google/martian/fifo"
	"github.com/google/martian/har"
	// side-effect importing to register with JSON API
	_ "github.com/google/martian/header"
	"github.com/google/martian/httpspec"
	mlog "github.com/google/martian/log"
	"github.com/google/martian/martianhttp"
	// side-effect importing to register with JSON API
	_ "github.com/google/martian/martianurl"
	_ "github.com/google/martian/method"
	"github.com/google/martian/mitm"
	// side-effect importing to register with JSON API
	_ "github.com/google/martian/pingback"
	_ "github.com/google/martian/priority"
	_ "github.com/google/martian/querystring"
	_ "github.com/google/martian/skip"
	_ "github.com/google/martian/status"
	"github.com/google/martian/verify"
)

// Martian is a wrapper for the initialized Martian proxy
type Martian struct {
	proxy    *martian.Proxy
	listener net.Listener
	mux      *http.ServeMux
}

// Start runs a martian.Proxy on addr
func Start(proxyAddr string) (*Martian, error) {
	return StartWithCertificate(proxyAddr, "", "")
}

// StartWithCertificate runs a proxy on addr and configures a cert for MITM
func StartWithCertificate(proxyAddr string, cert string, key string) (*Martian, error) {
	flag.Set("logtostderr", "true")

	signal.Ignore(syscall.SIGPIPE)

	l, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		return nil, err
	}

	mlog.Debugf("mobileproxy: started listener: %v", l.Addr())

	p := martian.NewProxy()

	mux := http.NewServeMux()
	p.SetMux(mux)

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

		mux.Handle("martian.proxy/authority.cer", martianhttp.NewAuthorityHandler(x509c))
		mlog.Debugf("mobileproxy: install cert from http://martian.proxy/authority.cer")
	}

	stack, fg := httpspec.NewStack("martian.mobileproxy")
	p.SetRequestModifier(stack)
	p.SetResponseModifier(stack)

	// add HAR logger
	hl := har.NewLogger()
	stack.AddRequestModifier(hl)
	stack.AddResponseModifier(hl)

	m := martianhttp.NewModifier()
	fg.AddRequestModifier(m)
	fg.AddResponseModifier(m)

	mlog.Debugf("mobileproxy: set martianhttp modifier")

	// Proxy specific handlers.
	// These handlers take precendence over proxy traffic and will not be intercepted.

	// Retrieve HAR logs
	mux.Handle("martian.proxy/logs", har.NewExportHandler(hl))
	mux.Handle("martian.proxy/logs/reset", har.NewResetHandler(hl))

	// Update modifiers.
	mux.Handle("martian.proxy/configure", m)
	mlog.Debugf("mobileproxy: configure with requests to http://martian.proxy/configure")

	// Verify assertions.
	vh := verify.NewHandler()
	vh.SetRequestVerifier(m)
	vh.SetResponseVerifier(m)
	mux.Handle("martian.proxy/verify", vh)
	mlog.Debugf("mobileproxy: check verifications with requests to http://martian.proxy/verify")

	// Reset verifications.
	rh := verify.NewResetHandler()
	rh.SetRequestVerifier(m)
	rh.SetResponseVerifier(m)
	mux.Handle("martian.proxy/verify/reset", rh)
	mlog.Debugf("mobileproxy: reset verifications with requests to http://martian.proxy/verify/reset")

	// Ignore SIGPIPE
	mlog.Debugf("mobileproxy: ignoring SIGPIPE for lldb")
	signal.Ignore(syscall.SIGPIPE)

	mlog.Infof("mobileproxy: starting proxy")
	go p.Serve(l)
	mlog.Infof("mobileproxy: started proxy on listener")

	return &Martian{
		proxy:    p,
		listener: l,
		mux:      mux,
	}, nil
}

// Shutdown tells the Proxy to close. The proxy will stay alive until all connections through it
// have closed or timed out.
func (p *Martian) Shutdown() {
	mlog.Infof("mobileproxy: telling proxy to close")
	p.proxy.Close()
	mlog.Infof("mobileproxy: proxy closed")
}

// Sets the Martian log level (Silent = 0, Error, Info, Debug), controlling which Martian
// log calls are displayed in the console
func SetLogLevel(l int) {
	mlog.SetLevel(l)
}
