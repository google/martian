// Copyright 2017 Google Inc. All rights reserved.
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

package mobile

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
	"github.com/google/martian/cors"
	"github.com/google/martian/cybervillains"
	"github.com/google/martian/fifo"
	"github.com/google/martian/har"
	"github.com/google/martian/httpspec"
	mlog "github.com/google/martian/log"
	"github.com/google/martian/marbl"
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

func init() {
	Init()
}

// Martian is a wrapper for the initialized Martian proxy
type Martian struct {
	proxy       *martian.Proxy
	listener    net.Listener
	mux         *http.ServeMux
	started     bool
	harLogging  bool
	TrafficPort int
	APIPort     int
	Cert        string
	Key         string
	AllowCORS   bool
}

// EnableCybervillains configures Martian to use the Cybervillians certificate.
func (m *Martian) EnableCybervillains() {
	m.Cert = cybervillains.Cert
	m.Key = cybervillains.Key
}

func (m *Martian) EnableHarLogging() {
	m.harLogging = true
}

// NewProxy creates a new Martian struct for configuring and starting a martian.
func NewProxy() *Martian {
	return &Martian{}
}

// Start starts the proxy given the configured values of the Martian struct.
func (m *Martian) Start() {
	var err error
	m.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", m.TrafficPort))
	if err != nil {
		log.Fatal(err)
	}

	mlog.Debugf("mobile: started listener on: %v", m.listener.Addr())
	m.proxy = martian.NewProxy()
	m.mux = http.NewServeMux()

	if m.Cert != "" && m.Key != "" {
		tlsc, err := tls.X509KeyPair([]byte(m.Cert), []byte(m.Key))
		if err != nil {
			log.Fatal(err)
		}

		mlog.Debugf("mobile: loaded cert and key")

		x509c, err := x509.ParseCertificate(tlsc.Certificate[0])
		if err != nil {
			log.Fatal(err)
		}

		mlog.Debugf("mobile: parsed cert")

		mc, err := mitm.NewConfig(x509c, tlsc.PrivateKey)
		if err != nil {
			log.Fatal(err)
		}

		mc.SetValidity(12 * time.Hour)
		mc.SetOrganization("Martian Proxy")

		m.proxy.SetMITM(mc)

		m.handle("/authority.cer", martianhttp.NewAuthorityHandler(x509c))
	}

	// Forward traffic that pattern matches in http.DefaultServeMux before applying
	// httpspec modifiers (via modifier, specifically)
	topg := fifo.NewGroup()
	apif := servemux.NewFilter(nil)
	apif.SetRequestModifier(api.NewForwarder("", m.APIPort))
	topg.AddRequestModifier(apif)

	stack, fg := httpspec.NewStack("martian.mobile")
	topg.AddRequestModifier(stack)
	topg.AddResponseModifier(stack)

	m.proxy.SetRequestModifier(topg)
	m.proxy.SetResponseModifier(topg)

	if m.harLogging {
		// add HAR logger for unmodified logs.
		uhl := har.NewLogger()
		fg.AddRequestModifier(uhl)
		fg.AddResponseModifier(uhl)

		// add HAR logger
		hl := har.NewLogger()
		stack.AddRequestModifier(hl)
		stack.AddResponseModifier(hl)

		// Retrieve Unmodified HAR logs
		m.handle("/logs/original", har.NewExportHandler(uhl))
		m.handle("/logs/original/reset", har.NewResetHandler(uhl))

		// Retrieve HAR logs
		m.handle("/logs", har.NewExportHandler(hl))
		m.handle("/logs/reset", har.NewResetHandler(hl))
	}

	lsh := marbl.NewHandler()
	// retrieve binary marbl logs
	m.handle("/binlogs", lsh)

	lsm := marbl.NewModifier(lsh)
	stack.AddRequestModifier(lsm)
	stack.AddResponseModifier(lsm)

	mod := martianhttp.NewModifier()
	fg.AddRequestModifier(mod)
	fg.AddResponseModifier(mod)

	// Proxy specific handlers.
	// These handlers take precendence over proxy traffic and will not be intercepted.

	// Update modifiers.
	m.handle("/configure", mod)

	// Verify assertions.
	vh := verify.NewHandler()
	vh.SetRequestVerifier(mod)
	vh.SetResponseVerifier(mod)

	m.handle("/verify", vh)

	// Reset verifications.
	rh := verify.NewResetHandler()
	rh.SetRequestVerifier(mod)
	rh.SetResponseVerifier(mod)
	m.handle("/verify/reset", rh)

	mlog.Infof("mobile: starting Martian proxy on listener")
	go m.proxy.Serve(m.listener)

	// start the API server
	apiAddr := fmt.Sprintf(":%d", m.APIPort)
	go http.ListenAndServe(apiAddr, m.mux)
	mlog.Infof("mobile: proxy API started on %s", apiAddr)
	m.started = true
}

// IsStarted returns true if the proxy has finished starting.
func (m *Martian) IsStarted() bool {
	return m.started
}

// Shutdown tells the Proxy to close. This function returns immediately, though
// there may still be connection threads hanging around until they time out
// depending on how the OS manages them.
func (m *Martian) Shutdown() {
	mlog.Infof("mobile: shutting down proxy")
	m.listener.Close()
	m.proxy.Close()
	m.started = false
	mlog.Infof("mobile: proxy shut down")
}

// SetLogLevel sets the Martian log level (Silent = 0, Error, Info, Debug), controlling which Martian
// log calls are displayed in the console
func SetLogLevel(l int) {
	mlog.SetLevel(l)
}

func (m *Martian) handle(pattern string, handler http.Handler) {
	if m.AllowCORS {
		handler = cors.NewHandler(handler)
	}
	m.mux.Handle(pattern, handler)
	mlog.Infof("mobile: handler registered for %s", pattern)

	lhp := path.Join(fmt.Sprintf("localhost:%d", m.APIPort), pattern)
	m.mux.Handle(lhp, handler)
	mlog.Infof("mobile: handler registered for %s", lhp)
}
