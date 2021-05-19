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

package testing

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/martian/v3/cybervillains"
	"github.com/google/martian/v3/mitm"
	"google.golang.org/grpc/credentials"
)

var (
	// CA is the certificate authority. It uses the Cybervillains key pair.
	CA *x509.Certificate
	// CAKey is the private key of the certificate authority.
	CAKey crypto.PrivateKey

	// RootCAs is a certificate pool containing `CA`.
	RootCAs *x509.CertPool

	// ClientTLS is a set of transport credentials to use with chains signed by `CA`.
	ClientTLS credentials.TransportCredentials

	// Localhost is a certificate for "localhost" signed by `CA`.
	Localhost *tls.Certificate
)

func init() {
	var err error
	CA, CAKey, err = initCA()
	if err != nil {
		log.Fatalf("Error initializing Cybervillains CA: %v", err)
	}

	RootCAs = x509.NewCertPool()
	RootCAs.AddCert(CA)
	ClientTLS = credentials.NewClientTLSFromCert(RootCAs, "")

	Localhost, err = initLocalhostCert(CA, CAKey)
	if err != nil {
		log.Fatalf("Error creating localhost server certificate: %v", err)
	}
}

func initCA() (*x509.Certificate, crypto.PrivateKey, error) {
	chain, err := tls.X509KeyPair([]byte(cybervillains.Cert), []byte(cybervillains.Key))
	if err != nil {
		return nil, nil, fmt.Errorf("creating Cybervillains root: %w", err)
	}
	cert, err := x509.ParseCertificate(chain.Certificate[0])
	if err != nil {
		return nil, nil, fmt.Errorf("parsing Cybervillains certificate: %w", err)
	}
	return cert, chain.PrivateKey, nil
}

func initLocalhostCert(ca *x509.Certificate, caPriv crypto.PrivateKey) (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating random key: %w", err)
	}

	// Subject Key Identifier support for end entity certificate.
	// https://www.ietf.org/rfc/rfc3280.txt (section 4.2.1.2)
	pkixpub, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, fmt.Errorf("marshalling public key: %w", err)
	}
	hasher := sha256.New()
	hasher.Write(pkixpub)
	keyID := hasher.Sum(nil)

	serial, err := rand.Int(rand.Reader, mitm.MaxSerialNumber)
	if err != nil {
		return nil, fmt.Errorf("generating serial number: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("getting hostname for creating cert: %w", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   hostname,
			Organization: []string{"Martian Proxy"},
		},
		SubjectKeyId:          keyID,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		DNSNames:              []string{hostname},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, priv.Public(), caPriv)
	if err != nil {
		return nil, fmt.Errorf("creating X509 server certificate: %w", err)
	}
	x509c, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, fmt.Errorf("parsing DER encoded certificate: %w", err)
	}
	return &tls.Certificate{
		Certificate: [][]byte{x509c.Raw, ca.Raw},
		PrivateKey:  priv,
		Leaf:        x509c,
	}, nil
}
