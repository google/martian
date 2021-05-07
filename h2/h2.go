// Copyright 2021 Google Inc. All rights reserved.
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

// Package h2 contains basic HTTP/2 handling for Martian.
package h2

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"sync"

	"github.com/google/martian/v3/log"
	"golang.org/x/net/http2"
)

var (
	// connectionPreface is the constant value of the connection preface.
	// https://tools.ietf.org/html/rfc7540#section-3.5
	connectionPreface = []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")
)

// Config stores the configuration information needed for HTTP/2 processing.
type Config struct {
	// AllowedHostsFilter is a function returning true if the argument is a host for which H2 is
	// permitted.
	AllowedHostsFilter func(string) bool

	// RootCAs is the pool of CA certificates used by the MitM client to authenticate the server.
	RootCAs *x509.CertPool

	// StreamProcessorFactories is a list of factories used to instantiate a chain of HTTP/2 stream
	// processors. A chain is created for every stream.
	StreamProcessorFactories []StreamProcessorFactory

	// EnableDebugLogs turns on fine-grained debug logging for HTTP/2.
	EnableDebugLogs bool
}

// Proxy proxies HTTP/2 traffic between a client connection, `cc`, and the HTTP/2 `url` assuming
// h2 is being used. Since no browsers use h2c, it's safe to assume all traffic uses TLS.
func (c *Config) Proxy(closing chan bool, cc io.ReadWriter, url *url.URL) error {
	if c.EnableDebugLogs {
		log.Infof("\u001b[1;35mProxying %v with HTTP/2\u001b[0m", url)
	}
	sc, err := tls.Dial("tcp", url.Host, &tls.Config{
		RootCAs:    c.RootCAs,
		NextProtos: []string{"h2"},
	})
	if err != nil {
		fmt.Printf("dial failed: %v\n", err)
		return fmt.Errorf("connecting h2 to %v: %w", url, err)
	}
	if err := forwardPreface(sc, cc); err != nil {
		return fmt.Errorf("initializing h2 with %v: %w", url, err)
	}

	cf, sf := http2.NewFramer(cc, cc), http2.NewFramer(sc, sc)
	cToS := newRelay(ClientToServer, "client", url.String(), cf, sf, &c.EnableDebugLogs)
	sToC := newRelay(ServerToClient, url.String(), "client", sf, cf, &c.EnableDebugLogs)

	// Completes circular parts of the initialization.

	// The client-to-server relay depends on the server-to-client relay and vice versa.
	cToS.peer, sToC.peer = sToC, cToS

	// Creating processors is circular because the create function references the relays and the
	// relays need to call create.
	cToS.processors = &streamProcessors{
		create: func(id uint32) *Processors {
			p := &Processors{cToS: &relayAdapter{id, cToS}, sToC: &relayAdapter{id, sToC}}
			// Chains the pipeline of processors together.
			for i := len(c.StreamProcessorFactories) - 1; i >= 0; i-- {
				cToS, sToC := c.StreamProcessorFactories[i](url, p)
				p = &Processors{
					cToS: newPartialProcessorAdapter(cToS, p.ForDirection(ClientToServer)),
					sToC: newPartialProcessorAdapter(sToC, p.ForDirection(ServerToClient)),
				}
			}
			return p
		},
	}
	sToC.processors = cToS.processors

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { // Forwards frames from client to server.
		defer wg.Done()
		if err := cToS.relayFrames(closing); err != nil {
			log.Errorf("relaying frame from client to %v: %v", url, err)
		}
	}()
	go func() { // Forwards frames from server to client.
		defer wg.Done()
		if err := sToC.relayFrames(closing); err != nil {
			log.Errorf("relaying frame from %v to client: %v", url, err)
		}
	}()
	wg.Wait()
	return nil
}

// forwardPreface forwards the connection preface from the client to the server.
func forwardPreface(server io.Writer, client io.Reader) error {
	preface := make([]byte, len(connectionPreface))
	if _, err := client.Read(preface); err != nil {
		return fmt.Errorf("reading preface: %w", err)
	}
	if !bytes.Equal(preface, connectionPreface) {
		return fmt.Errorf("client sent unexpected preface: %s", hex.Dump(preface))
	}
	for m := len(connectionPreface); m > 0; {
		n, err := server.Write([]byte(preface))
		if err != nil {
			return fmt.Errorf("writing preface: %w", err)
		}
		preface = preface[n:]
		m -= n
	}
	return nil
}

type streamProcessors struct {
	// processors stores `*Processors` instances keyed by uint32 stream ID.
	processors sync.Map

	// create creates `*Processors` for the given stream ID.
	create func(uint32) *Processors
}

// Get returns a the processor with the given ID and direction.
func (s *streamProcessors) Get(id uint32, dir Direction) Processor {
	value, ok := s.processors.Load(id)
	if !ok {
		value, _ = s.processors.LoadOrStore(id, s.create(id))
	}
	return value.(*Processors).ForDirection(dir)
}
