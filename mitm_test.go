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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var mitm *MITM

func init() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	start, end := time.Now().Add(-5*time.Hour), time.Now().Add(5*time.Hour)
	tpl, err := NewTemplate("authority.test", "Authority", start, end, &priv.PublicKey)
	if err != nil {
		panic(err)
	}
	tpl.IsCA = true
	tpl.KeyUsage |= x509.KeyUsageCertSign

	cb, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	authority, err := x509.ParseCertificate(cb)
	if err != nil {
		panic(err)
	}

	mitm = &MITM{
		Authority:	authority,
		PublicKey:	&priv.PublicKey,
		PrivateKey:	priv,
		Organization:	"Organization",
		Validity:	time.Hour,
	}
}

func TestMITMHijack(t *testing.T) {
	p1, p2 := net.Pipe()
	defer p1.Close()
	defer p2.Close()

	_, rw, err := mitm.Hijack(p1, "example.com")
	if err != nil {
		t.Fatalf("mitm.Hijack(r, %q): got %v, want no error", "example.com", err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(mitm.Authority)

	client := tls.Client(p2, &tls.Config{
		ServerName:	"example.com",
		RootCAs:	pool,
	})

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q, %q, nil): got %v, want no error", "GET", "https://example.com", err)
	}

	go req.WriteProxy(client)

	req2, err := http.ReadRequest(rw.Reader)
	if err != nil {
		t.Fatalf("http.ReadRequest(rw.Reader): got %v, want no error", err)
	}

	if got, want := req2.URL.String(), req.URL.String(); got != want {
		t.Errorf("req2.URL: got %q, want %q", got, want)
	}
}

func TestMITMNewTemplate(t *testing.T) {
	start, end := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)

	tpl, err := NewTemplate("Organization", "example.com", start, end, mitm.PublicKey)
	if err != nil {
		t.Fatalf("NewTemplate(): got %v, want no error", err)
	}

	if got := tpl.SerialNumber; got.Cmp(MaxSerialNumber) >= 0 {
		t.Errorf("tpl.SerialNumber: got %v, want <= MaxSerialNumber", got)
	}

	if got, want := tpl.Subject.CommonName, "example.com"; got != want {
		t.Errorf("tpl.Subject.CommonName: got %q, want %q", got, want)
	}
	if got, want := tpl.Subject.Organization, []string{"Organization"}; !reflect.DeepEqual(got, want) {
		t.Errorf("tpl.Subject.Organization: got %v, want %v", got, want)
	}

	if got := tpl.SubjectKeyId; got == nil {
		t.Error("tpl.SubjectKeyId: got nil, want to be present")
	}

	if got, want := tpl.KeyUsage, x509.KeyUsageKeyEncipherment; got&want == 0 {
		t.Error("tpl.KeyUsage: got nothing, want to include x509.KeyUsageKeyEncipherment")
	}
	if got, want := tpl.KeyUsage, x509.KeyUsageDigitalSignature; got&want == 0 {
		t.Error("tpl.KeyUsage: got nothing, want to include x509.KeyUsageDigitalSignature")
	}

	want := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	if got := tpl.ExtKeyUsage; !reflect.DeepEqual(got, want) {
		t.Errorf("tpl.ExtKeyUsage: got %v, want %v", got, want)
	}

	if !tpl.BasicConstraintsValid {
		t.Error("tpl.BasicConstraintsValid: got false, want true")
	}

	if got, want := tpl.DNSNames, []string{"example.com"}; !reflect.DeepEqual(got, want) {
		t.Errorf("tpl.DNSNames: got %v, want %v", got, want)
	}

	if got, want := tpl.NotBefore, start; !got.Equal(want) {
		t.Errorf("tpl.NotBefore: got %v, want %v", got, want)
	}
	if got, want := tpl.NotAfter, end; !got.Equal(want) {
		t.Errorf("tpl.NotAfter: got %v, want %v", got, want)
	}
}
