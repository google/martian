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

package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/google/martian"
	"github.com/google/martian/api"
	"github.com/google/martian/cors"
	"github.com/google/martian/fifo"
	"github.com/google/martian/httpspec"
	"github.com/google/martian/mitm"
	"github.com/google/martian/servemux"
	"github.com/google/martian/trafficshape"
)

// Server is a Martian Proxy server.
type Server struct {
	Proxy *martian.Proxy

	name            string
	trafficPort     int
	apiPort         int
	trafficListener net.Listener
	apiListener     net.Listener
	mitmCert        string
	mitmKey         string
	mux             *http.ServeMux
	roundtripper    http.RoundTripper
	allowCORS       bool
	trafficShaping  bool
	apiCertPath     string
	apiKeyPath      string
	downstreamProxy *url.URL
	modifiers       *fifo.Group
}

// Start starts the proxy server. Blocks until SIGINT or SIGKILL is received.
func (s *Server) Start() error {
	tfcl, err := net.Listen("tcp", fmt.Sprintf(":%d", s.trafficPort))
	if err != nil {
		return err
	}
	s.trafficListener = tfcl

	if s.trafficShaping {
		tsl := trafficshape.NewListener(tfcl)
		tsh := trafficshape.NewHandler(tsl)
		s.handle("/shape-traffic", tsh)
		s.trafficListener = tsl
	}

	s.Proxy.SetRequestModifier(s.modifiers)
	s.Proxy.SetResponseModifier(s.modifiers)

	apil, err := net.Listen("tcp", fmt.Sprintf(":%d", s.apiPort))
	if err != nil {
		return err
	}

	s.apiListener = apil

	go s.Proxy.Serve(s.trafficListener)

	if s.apiCertPath != "" && s.apiKeyPath != "" {
		go http.ServeTLS(s.apiListener, s.mux, s.apiCertPath, s.apiKeyPath)
	} else {
		go http.Serve(s.apiListener, s.mux)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	<-sigc

	return nil
}

// NewServer returns a Server.
func NewServer(name string, trafficPort, apiPort int, options ...func(*Server) error) (*Server, error) {
	svr := &Server{
		name:        name,
		trafficPort: trafficPort,
		apiPort:     apiPort,
		Proxy:       martian.NewProxy(),
		modifiers:   fifo.NewGroup(),
	}

	// Forward api traffic that matches in svr.mux before httpspec and logging modifiers
	apif := servemux.NewFilter(svr.mux)
	apif.SetRequestModifier(api.NewForwarder("", svr.apiPort))
	svr.modifiers.AddRequestModifier(apif)

	stack, _ := httpspec.NewStack(name)
	svr.modifiers.AddRequestModifier(stack)
	svr.modifiers.AddResponseModifier(stack)

	for _, optionFunc := range options {
		optionFunc(svr)
	}

	return svr, nil
}

// SetRoundTripper sets the proxy HTTP roundtripper.
func SetRoundTripper(rt http.RoundTripper) func(*Server) error {
	return func(s *Server) error {
		s.Proxy.SetRoundTripper(rt)

		return nil
	}
}

// EnableMITM enables man-in-the-middle with the provided certificate and private key.
func EnableMITM(cert, key string) func(*Server) error {
	return func(s *Server) error {
		return s.enableMITM(cert, key)
	}
}

func (s *Server) enableMITM(cert, key string) error {
	tlsc, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return err
	}

	x509c, err := x509.ParseCertificate(tlsc.Certificate[0])
	if err != nil {
		return err
	}

	mc, err := mitm.NewConfig(x509c, tlsc.PrivateKey)
	if err != nil {
		return err
	}

	mc.SetValidity(12 * time.Hour)
	mc.SetOrganization(s.name)

	s.Proxy.SetMITM(mc)

	return nil
}

// APIOverTLS enables TLS for API requests.
func APIOverTLS(certPath, keyPath string) func(*Server) error {
	return func(s *Server) error {
		s.apiCertPath = certPath
		s.apiKeyPath = keyPath

		return nil
	}
}

// EnableTrafficShaping enables traffic shaping.
func EnableTrafficShaping() func(*Server) error {
	return func(s *Server) error {
		s.trafficShaping = true

		return nil
	}
}

// AllowCORS allows CORS requests for the API.
func AllowCORS() func(*Server) error {
	return func(s *Server) error {
		s.allowCORS = true

		return nil
	}
}

// DownstreamProxy sets the url of a proxy that requests are forwarded to.
func DownstreamProxy(downstreamProxy *url.URL) func(*Server) error {
	return func(s *Server) error {
		s.Proxy.SetDownstreamProxy(downstreamProxy)

		return nil
	}
}

// AddModifiers sets the modifiers.
func AddModifiers(mods martian.RequestResponseModifier, setPath, resetPath string) func(*Server) error {
	return func(s *Server) error {

		return nil
	}
}

// SetPremodificationLogger sets logging before Runtime Configurable modifiers.
func SetPremodificationLogger(logger martian.RequestResponseModifier,
	handlers map[string]func(martian.RequestResponseModifier) http.HandlerFunc) func(*Server) error {

	return func(s *Server) error {
		smuxf := servemux.NewFilter(s.mux)
		smuxf.RequestWhenFalse(logger)
		smuxf.ResponseWhenFalse(logger)

		for pattern, handler := range handlers {
			s.handlerFunc(pattern, handler(logger))
		}

		return nil
	}
}

func (s *Server) handle(pattern string, handler http.Handler) {
	if s.allowCORS {
		handler = cors.NewHandler(handler)
	}

	s.mux.Handle(pattern, handler)

	lhp := path.Join(fmt.Sprintf("localhost:%d", s.apiPort), pattern)
	s.mux.Handle(lhp, handler)
}

func (s *Server) handlerFunc(pattern string, handleFunc http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handleFunc)

	lhp := path.Join(fmt.Sprintf("localhost:%d", s.apiPort), pattern)
	s.mux.HandleFunc(lhp, handleFunc)
}

type loggerFunc = func(martian.RequestResponseModifier) http.HandlerFunc
