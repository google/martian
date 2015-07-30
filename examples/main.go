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

/*
proxy is a martian.Proxy configurable via HTTP.

It can be dynamically configured/queried at runtime by issuing requests to
proxy specific paths using JSON.

Supported configuration endpoints:

	POST http://martian.proxy/configure

sets the request and response modifier of the proxy; modifiers adhere to the
following top-level JSON structure:

	{
		"package.Modifier": {
			"scope": ["request", "response"],
			"attribute 1": "value",
			"attribute 2": "value"
		}
	}

modifiers may be "stacked" to provide support for additional behaviors; for
example, to add a "Martian-Test" header with the value "true" for requests
with the domain "www.example.com" the JSON message would be:

	{
		"url.Filter": {
			"scope": ["request"],
			"host": "www.example.com",
			"modifier": {
				"header.Modifier": {
					"name": "Martian-Test",
					"value": "true"
				}
			}
		}
	}

url.Filter parses the JSON object in the value of the "url.Filter" attribute;
the "host" key tells the url.Filter to filter requests if the host explicitly
matches "www.example.com"

the "modifier" key within the "url.Filter" JSON object contains another
modifier message of the type header.Modifier to run iff the filter passes

groups may also be used to run multiple modifiers sequentially; for example to
log requests and responses after adding the "Martian-Test" header to the
request, but only when the host matches www.example.com:

  {
    "url.Filter": {
      "host": "www.example.com",
      "modifier": {
        "fifo.Group": {
          "modifiers": [
            {
              "header.Modifier": {
                "scope": ["request"],
                "name": "Martian-Test",
                "value": "true"
              }
            },
            {
              "log.Logger": { }
            }
          ]
        }
      }
    }
  }

modifiers are designed to be composed together in ways that allow the user to
write a single JSON structure to accomplish a variety of functionality

	GET http://martian.proxy/verify

retrieves the verifications errors as JSON with the following structure:

	{
		"errors": [
			{
				"message": "request(url) verification failure"
			},
			{
				"message": "response(url) verification failure"
			}
		]
	}

verifiers also adhere to the modifier interface and thus can be included in the
modifier configuration request; for example, to verify that all requests to
"www.example.com" are sent over HTTPS send the following JSON to the
configuration endpoint:

	{
		"url.Filter": {
			"scope": ["request"],
			"host": "www.example.com",
			"modifier": {
				"url.Verifier": {
					"scope": ["request"],
					"scheme": "https"
				}
			}
		}
	}

sending a request to "http://martian.proxy/verify" will then return errors from the url.Verifier

	POST http://martian.proxy/verify/reset

resets the verifiers to their initial state; note some verifiers may start in
a failure state (e.g., pingback.Verifier is failed if no requests have been
seen by the proxy)

	GET http://martian.proxy/authority.cer

prompts the user to install the CA certificate used by the proxy if MITM is enabled

passing the -cors flag will enable CORS support for the endpoints so that they
may be called via AJAX

The flags are:
	-addr=":8080"
		host:port of the proxy
	-api-addr=":0"
		host:port of the configuration API
	-api-hostname="martian.proxy"
		hostname that can be used to reference the configuration API when
		configuring through the proxy
	-cert=""
		PEM encoded X.509 CA certificate; if set, it will be set as the
		issuer for dynamically-generated certificates during man-in-the-middle
	-key=""
		PEM encoded private key of cert (RSA or ECDSA); if set, the key will be used
		to sign dynamically-generated certificates during man-in-the-middle
	-generate-cert=false
		generates a CA certificate and private key to use for man-in-the-middle;
		the certificate is only valid while the proxy is running and will be
		discarded on shutdown
	-organization="Martian Proxy"
		organization name set on the dynamically-generated certificates during
		man-in-the-middle
	-validity="1h"
		window of time around the time of request that the dynamically-generated
		certificate is valid for; the duration is set such that the total valid
		timeframe is double the value of validity (1h before & 1h after)
	-cors=false
		allow the proxy to be configured via CORS requests; such as when
		configuring the proxy via AJAX
*/
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/martian"
	"github.com/google/martian/cors"
	"github.com/google/martian/fifo"
	"github.com/google/martian/header"
	"github.com/google/martian/martianhttp"
	"github.com/google/martian/martianurl"
	"github.com/google/martian/mitm"
	"github.com/google/martian/verify"

	_ "github.com/google/martian/body"
	_ "github.com/google/martian/cookie"
	_ "github.com/google/martian/log"
	_ "github.com/google/martian/method"
	_ "github.com/google/martian/pingback"
	_ "github.com/google/martian/priority"
	_ "github.com/google/martian/querystring"
	_ "github.com/google/martian/status"
)

var (
	addr         = flag.String("addr", ":8080", "host:port of the proxy")
	apiAddr      = flag.String("api-addr", ":0", "host:port of the configuration API")
	apiHostname  = flag.String("api-hostname", "martian.proxy", "hostname forwarded to apiAddr")
	generateCert = flag.Bool("generate-cert", false, "generate certificate and private key for MITM")
	cert         = flag.String("cert", "", "CA certificate used to sign MITM certificates")
	key          = flag.String("key", "", "private key of the CA used to sign MITM certificates")
	organization = flag.String("organization", "Martian Proxy", "organization name for MITM certificates")
	validity     = flag.Duration("validity", time.Hour, "window of time that MITM certificates are valid")
	allowCORS    = flag.Bool("cors", false, "allow CORS requests to configure the proxy")
)

func main() {
	flag.Parse()

	p := martian.NewProxy()

	var x509c *x509.Certificate
	var priv interface{}

	if *generateCert {
		var err error
		x509c, priv, err = mitm.NewAuthority("martian.proxy", "Martian Authority", 30*24*time.Hour)
		if err != nil {
			log.Fatal(err)
		}
	} else if *cert != "" && *key != "" {
		tlsc, err := tls.LoadX509KeyPair(*cert, *key)
		if err != nil {
			log.Fatal(err)
		}
		priv = tlsc.PrivateKey

		x509c, err = x509.ParseCertificate(tlsc.Certificate[0])
		if err != nil {
			log.Fatal(err)
		}
	}

	if x509c != nil && priv != nil {
		mc, err := mitm.NewConfig(x509c, priv)
		if err != nil {
			log.Fatal(err)
		}

		mc.SetValidity(*validity)
		mc.SetOrganization(*organization)

		p.SetMITM(mc)

		// Expose certificate authority.
		ah := martianhttp.NewAuthorityHandler(x509c)
		http.Handle("/authority.cer", ah)
	}

	fg := fifo.NewGroup()

	hbhmod := header.NewHopByHopModifier()
	fg.AddRequestModifier(hbhmod)
	fg.AddRequestModifier(header.NewForwardedModifier())
	fg.AddRequestModifier(header.NewBadFramingModifier())
	fg.AddRequestModifier(header.NewViaModifier("martian 1.1"))

	m := martianhttp.NewModifier()
	fg.AddRequestModifier(m)
	fg.AddResponseModifier(m)

	fg.AddResponseModifier(hbhmod)

	p.SetRequestModifier(fg)
	p.SetResponseModifier(fg)

	// Proxy specific handlers.
	// These handlers take precendence over proxy traffic and will not be
	// intercepted.

	// Configure modifiers.
	configure("/configure", m)

	// Verify assertions.
	vh := verify.NewHandler()
	vh.SetRequestVerifier(m)
	vh.SetResponseVerifier(m)
	configure("/verify", vh)

	// Reset verifications.
	rh := verify.NewResetHandler()
	rh.SetRequestVerifier(m)
	rh.SetResponseVerifier(m)
	configure("/verify/reset", rh)

	cl, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Redirect *apiHostname to the API server host:port.
	apif := martianurl.NewFilter(&url.URL{
		Host: *apiHostname,
	})
	apim := martianurl.NewModifier(&url.URL{
		Host: cl.Addr().String(),
	})
	apif.SetRequestModifier(apim)

	fg.AddRequestModifier(apif)

	go http.Serve(cl, nil)

	log.Printf("martian: API started at %s\n", cl.Addr())

	pl, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("martian: proxy started at %s\n", pl.Addr())

	log.Fatal(p.Serve(pl))
}

// configure installs a configuration handler at path.
func configure(path string, handler http.Handler) {
	if *allowCORS {
		handler = cors.NewHandler(handler)
	}

	http.Handle(path, handler)
}
