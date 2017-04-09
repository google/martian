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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/martian/mitm"
)

func waitForProxy(t *testing.T, c *http.Client, apiUrl string) {
	timeout := 5 * time.Second
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		res, err := c.Get(apiUrl)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		defer res.Body.Close()
		if got, want := res.StatusCode, http.StatusOK; got != want {
			t.Fatalf("waitForProxy: c.Get(%q): got status %d, want %d", apiUrl, got, want)
		}
		return
	}
	t.Fatalf("waitForProxy: did not start up within %.1f seconds", timeout.Seconds())
}

// getFreePort returns a port string preceded by a colon, e.g. ":1234"
func getFreePort(t *testing.T) string {
	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("getFreePort: could not get free port: %v", err)
	}
	defer l.Close()
	return l.Addr().String()[strings.LastIndex(l.Addr().String(), ":"):]
}

func parseURL(t *testing.T, u string) *url.URL {
	p, err := url.Parse(u)
	if err != nil {
		t.Fatalf("url.Parse(%q): got error %v, want no error", u, err)
	}
	return p
}

func TestProxy(t *testing.T) {
	tempDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Build proxy binary
	binPath := filepath.Join(tempDir, "proxy")
	cmd := exec.Command("go", "build", "-o", binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	t.Run("Http", func(t *testing.T) {
		// Start proxy
		proxyPort := getFreePort(t)
		apiPort := getFreePort(t)
		cmd = exec.Command(binPath, "-addr="+proxyPort, "-api-addr="+apiPort)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer cmd.Wait()
		defer cmd.Process.Signal(os.Interrupt)

		proxyUrl := "http://localhost" + proxyPort
		apiUrl := "http://localhost" + apiPort
		configureUrl := "http://martian.proxy/configure"

		// TODO: Make using API hostport directly work on Travis.
		apiClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parseURL(t, apiUrl))}}
		waitForProxy(t, apiClient, configureUrl)

		// Configure modifiers
		config := strings.NewReader(`
			{
			  "fifo.Group": {
			    "scope": ["request", "response"],
			    "modifiers": [
			      {
			        "status.Modifier": {
			          "scope": ["response"],
			          "statusCode": 418
			        }
			      },
			      {
			        "skip.RoundTrip": {}
			      }
			    ]
			  }
			}`)
		res, err := apiClient.Post(configureUrl, "application/json", config)
		if err != nil {
			t.Fatalf("apiClient.Post(%q): got error %v, want no error", configureUrl, err)
		}
		defer res.Body.Close()
		if got, want := res.StatusCode, http.StatusOK; got != want {
			t.Fatalf("apiClient.Post(%q): got status %d, want %d", configureUrl, got, want)
		}

		// Exercise proxy
		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parseURL(t, proxyUrl))}}

		testUrl := "http://super.fake.domain/"
		res, err = client.Get(testUrl)
		if err != nil {
			t.Fatalf("client.Get(%q): got error %v, want no error", testUrl, err)
		}
		defer res.Body.Close()
		if got, want := res.StatusCode, http.StatusTeapot; got != want {
			t.Errorf("client.Get(%q): got status %d, want %d", testUrl, got, want)
		}
	})

	t.Run("HttpsGenerateCert", func(t *testing.T) {
		// Create test certificate for test TLS server
		certName := "martian.proxy"
		certOrg := "Martian Authority"
		certExpiry := 90 * time.Minute
		servCert, servPriv, err := mitm.NewAuthority(certName, certOrg, certExpiry)
		if err != nil {
			t.Fatalf("mitm.NewAuthority(%q, %q, %q): got error %v, want no error", certName, certOrg, certExpiry, err)
		}
		mc, err := mitm.NewConfig(servCert, servPriv)
		if err != nil {
			t.Fatalf("mitm.NewConfig(%p, %q): got error %v, want no error", servCert, servPriv, err)
		}
		sc := mc.TLS()

		// Configure and start test TLS server
		servPort := getFreePort(t)
		l, err := tls.Listen("tcp", servPort, sc)
		if err != nil {
			t.Fatalf("tls.Listen(\"tcp\", %q, %p): got error %v, want no error", servPort, sc, err)
		}
		defer l.Close()

		server := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("Hello!"))
			}),
		}
		go server.Serve(l)
		defer server.Close()

		// Start proxy
		proxyPort := getFreePort(t)
		apiPort := getFreePort(t)
		cmd = exec.Command(binPath, "-addr="+proxyPort, "-api-addr="+apiPort, "-generate-ca-cert", "-skip-tls-verify")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer cmd.Wait()
		defer cmd.Process.Signal(os.Interrupt)

		proxyUrl := "http://localhost" + proxyPort
		apiUrl := "http://localhost" + apiPort
		configureUrl := "http://martian.proxy/configure"

		// TODO: Make using API hostport directly work on Travis.
		apiClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parseURL(t, apiUrl))}}
		waitForProxy(t, apiClient, configureUrl)

		// Configure modifiers
		config := strings.NewReader(fmt.Sprintf(`
			{
			  "body.Modifier": {
			    "scope": ["response"],
			    "contentType": "text/plain",
			    "body": "%s"
			  }
			}`, base64.StdEncoding.EncodeToString([]byte("茶壺"))))
		res, err := apiClient.Post(configureUrl, "application/json", config)
		if err != nil {
			t.Fatalf("apiClient.Post(%q): got error %v, want no error", configureUrl, err)
		}
		defer res.Body.Close()
		if got, want := res.StatusCode, http.StatusOK; got != want {
			t.Fatalf("apiClient.Post(%q): got status %d, want %d", configureUrl, got, want)
		}

		// Install proxy's CA cert into http client
		caCertUrl := "http://martian.proxy/authority.cer"
		res, err = apiClient.Get(caCertUrl)
		if err != nil {
			t.Fatalf("apiClient.Get(%q): got error %v, want no error", caCertUrl, err)
		}
		defer res.Body.Close()
		caCert, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("ioutil.ReadAll(res.Body): got error %v, want no error", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Exercise proxy
		client := &http.Client{Transport: &http.Transport{
			Proxy: http.ProxyURL(parseURL(t, proxyUrl)),
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}}

		testUrl := "https://localhost" + servPort
		res, err = client.Get(testUrl)
		if err != nil {
			t.Fatalf("client.Get(%q): got error %v, want no error", testUrl, err)
		}
		defer res.Body.Close()
		if got, want := res.StatusCode, http.StatusTeapot; got != want {
			t.Fatalf("client.Get(%q): got status %d, want %d", testUrl, got, want)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("ioutil.ReadAll(res.Body): got error %v, want no error", err)
		}
		if got, want := string(body), "茶壺"; got != want {
			t.Fatalf("modified response body: got %s, want %s", got, want)
		}
	})
}
