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
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

// MaxSerialNumber is the upper boundary that is used to create unique serial
// numbers for the certificate. This can be any unsigned integer up to 20
// bytes (2^(8*20)-1).
var MaxSerialNumber = big.NewInt(0).SetBytes(bytes.Repeat([]byte{255}, 20))

// MITM is the configuration for using the Proxy as a MITM.
type MITM struct {
	// Authority is the CA certificate used to sign MITM certificates.
	Authority	*x509.Certificate
	// PublicKey used to create MITM certificates.
	PublicKey	interface{}
	// PrivateKey of the CA used to sign MITM certificates.
	PrivateKey	interface{}
	// Validity is the window of time around time.Now() that the
	// certificate will be valid.
	Validity	time.Duration
	// Organization that is displayed as the owner of the
	// certificate.
	Organization	string
}

// Hijack takes a net.Conn and the host name to create the SSL
// certificate for and returns a tls.Conn that can read and write
// to the given host over TLS.
func (mitm *MITM) Hijack(conn net.Conn, host string) (*tls.Conn, *bufio.ReadWriter, error) {
	// Ensure the certificate we create is valid within a window of time to allow
	// for clock skew.
	start := time.Now().Add(-mitm.Validity)
	end := time.Now().Add(mitm.Validity)

	tpl, err := NewTemplate(mitm.Organization, host, start, end, mitm.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	cb, err := x509.CreateCertificate(rand.Reader, tpl, mitm.Authority, mitm.PublicKey, mitm.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{
			{
				PrivateKey:	mitm.PrivateKey,
				Certificate:	[][]byte{cb},
			},
		},
	}

	tlsConn := tls.Server(conn, config)
	r := bufio.NewReader(tlsConn)
	w := bufio.NewWriter(tlsConn)

	return tlsConn, bufio.NewReadWriter(r, w), nil
}

// NewTemplate returns a new base *x509.Certificate.
func NewTemplate(org, host string, start, end time.Time, pub interface{}) (*x509.Certificate, error) {
	pkixPub, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	h.Write(pkixPub)
	keyID := h.Sum(nil)

	serial, err := rand.Int(rand.Reader, MaxSerialNumber)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber:	serial,
		Subject: pkix.Name{
			CommonName:	host,
			Organization:	[]string{org},
		},
		SubjectKeyId:		keyID,
		KeyUsage:		x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid:	true,
		DNSNames:		[]string{host},
		NotBefore:		start,
		NotAfter:		end,
	}, nil
}
