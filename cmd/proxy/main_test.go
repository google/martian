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

func TestProxyHttp(t *testing.T) {
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

	au, err := url.Parse(apiUrl)
	if err != nil {
		t.Fatalf("url.Parse(%q): got error %v, want no error", apiUrl, err)
	}
	// TODO: Make using API hostport directly work on Travis.
	apiClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(au)}}
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
	pu, err := url.Parse(proxyUrl)
	if err != nil {
		t.Fatalf("url.Parse(%q): got error %v, want no error", proxyUrl, err)
	}
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}

	testUrl := "http://super.fake.domain/"
	res, err = client.Get(testUrl)
	if err != nil {
		t.Fatalf("client.Get(%q): got error %v, want no error", testUrl, err)
	}
	defer res.Body.Close()
	if got, want := res.StatusCode, http.StatusTeapot; got != want {
		t.Errorf("client.Get(%q): got status %d, want %d", testUrl, got, want)
	}
}
