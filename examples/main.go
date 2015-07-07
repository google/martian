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

	POST host:port/martian/modifiers

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

	GET host:port/martian/verify

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

sending a request to "host:port/martian/verify" will then return errors from the url.Verifier

	POST host:port/martian/verify/reset

resets the verifiers to their initial state; note some verifiers may start in
a failure state (e.g., pingback.Verifier is failed if no requests have been
seen by the proxy)

passing the -cors flag will enable CORS support for the endpoints so that they
may be called via AJAX

The flags are:
	-addr=":8080"
		host:port of the proxy
	-cert=""
		PEM encoded X509 CA certificate; if set, it will be set as the
		issuer for dynamically-generated certificates during man-in-the-middle
	-key=""
		PEM encoded private key of cert (RSA or ECDSA); if set, the key will be used
		to sign dynamically-generated certificates during man-in-the-middle
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
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/google/martian"
	"github.com/google/martian/cors"
	"github.com/google/martian/fifo"
	"github.com/google/martian/header"
	"github.com/google/martian/martianhttp"
	"github.com/google/martian/verify"
)

var (
	addr         = flag.String("addr", ":8080", "host:port of the proxy")
	cert         = flag.String("cert", "", "CA certificate used to sign MITM certificates")
	key          = flag.String("key", "", "private key of the CA used to sign MITM certificates")
	organization = flag.String("organization", "Martian Proxy", "organization name for MITM certificates")
	validity     = flag.Duration("validity", time.Hour, "window of time that MITM certificates are valid")
	allowCORS    = flag.Bool("cors", false, "allow CORS requests to configure the proxy")
)

func main() {
	flag.Parse()

	var mitm *martian.MITM
	if *cert != "" && *key != "" {
		tlsc, err := tls.LoadX509KeyPair(*cert, *key)
		if err != nil {
			log.Fatal(err)
		}

		x509c, err := x509.ParseCertificate(tlsc.Certificate[0])
		if err != nil {
			log.Fatal(err)
		}

		var pub crypto.PublicKey
		switch priv := tlsc.PrivateKey.(type) {
		case *rsa.PrivateKey:
			pub = &priv.PublicKey
		case *ecdsa.PrivateKey:
			pub = &priv.PublicKey
		default:
			log.Fatal("Public key is not of supported type: rsa, ecdsa.")
		}

		mitm = &martian.MITM{
			Authority:    x509c,
			PublicKey:    pub,
			PrivateKey:   tlsc.PrivateKey,
			Validity:     *validity,
			Organization: *organization,
		}
	}

	p := martian.NewProxy(mitm)

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

	// Update modifiers.
	configure("martian/modifiers", m)

	// Verify assertions.
	vh := verify.NewHandler()
	vh.SetRequestVerifier(m)
	vh.SetResponseVerifier(m)
	configure("martian/verify", vh)

	// Reset verifications.
	rh := verify.NewResetHandler()
	rh.SetRequestVerifier(m)
	rh.SetResponseVerifier(m)
	configure("martian/verify/reset", rh)

	// Forward all other requests to the proxy.
	http.Handle("/", p)

	log.Printf("Martian started at %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// configure installs a configuration handler at path.
func configure(path string, handler http.Handler) {
	if *allowCORS {
		handler = cors.NewHandler(handler)
	}

	http.Handle(path, handler)
}
