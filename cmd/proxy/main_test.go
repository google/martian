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

const (
	binName = "proxy"
	timeout = 5 * time.Second
)

func waitForProxyLive(t *testing.T, c *http.Client) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		res, err := c.Get("http://martian.proxy/configure")
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if got, want := res.StatusCode, http.StatusOK; got != want {
			t.Fatalf("GET config: got status code %d, want %d", got, want)
		}
		return
	}
	t.Fatalf("Proxy did not start up within %v seconds", timeout.Seconds())
}

func getProxiedClient(t *testing.T, proxyUrl string) *http.Client {
	pu, err := url.Parse(proxyUrl)
	if err != nil {
		t.Fatalf("Parse proxy url: got error %v, want no error", err)
	}
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
}

// getFreePort returns a port string preceded by a colon, e.g. ":1234"
func getFreePort(t *testing.T) string {
	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("Could not get free port: %v", err)
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

	binPath := filepath.Join(tempDir, binName)

	// Build proxy binary
	cmd := exec.Command("go", "build", "-o", binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Start proxy
	proxyPort := getFreePort(t)
	apiPort := getFreePort(t)
	t.Logf("proxyPort=%s apiPort=%s", proxyPort, apiPort)
	cmd = exec.Command(binPath, "-addr="+proxyPort, "-api-addr="+apiPort)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()
	defer cmd.Process.Signal(os.Interrupt)

	apiClient := getProxiedClient(t, "http://localhost"+apiPort)
	waitForProxyLive(t, apiClient)

	// Configure modifiers
	configReader := strings.NewReader(`
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
	res, err := apiClient.Post("http://martian.proxy/configure", "application/json", configReader)
	if err != nil {
		t.Fatalf("POST config: got error %v, want no error", err)
	}
	if got, want := res.StatusCode, http.StatusOK; got != want {
		t.Fatalf("POST config: got status code %d, want %d", got, want)
	}

	// Exercise proxy
	client := getProxiedClient(t, "http://localhost"+proxyPort)
	res, err = client.Get("http://super.fake.domain/")
	if err != nil {
		t.Fatalf("GET request: got error %v, want no error", err)
	}
	if got, want := res.StatusCode, http.StatusTeapot; got != want {
		t.Errorf("GET request: got status code %d, want %d", got, want)
	}
}
