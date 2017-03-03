package martiantest

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

type CounterServer int

const (
	RootDir = "/home/bighead/Code/go/src/github.com/google/martian"
)

func (cs *CounterServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%d", *cs)
	(*cs) += 1
}

func fetchValueByRequest(u, proxy string) (int, error) {
	var c *http.Client
	if proxy != "" {
		pu, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		c = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
	} else {
		c = http.DefaultClient
	}

	r, err := c.Get(u)
	if err != nil {
		return 0, err
	}
	log.Print(r.Header)
	ct, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	if v, err := strconv.Atoi(string(ct)); err != nil {
		return 0, err
	} else {
		return v, nil
	}
}

func PostJsonConfigToMartian(u, j string) {
	for {
		pu, err := url.Parse(u)
		c := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
		_, err = c.Post("http://martian.proxy/configure", "application/json", strings.NewReader(j))
		if err != nil {
			log.Printf("Problem configuring martian %v", err)
			time.Sleep(time.Second)
		} else {
			return
		}
	}

}

func TestUsingCachedResponset(t *testing.T) {
	// First bring up martian: go run cmd/proxy/main.go -addr :8889 -api-addr :8890

	mch := make(chan int)
	go func(pt, apt int) {
		log.Printf("Building martian")
		if err := exec.Command("go", "build", path.Join(RootDir, "cmd/proxy/main.go")).Run(); err != nil {
			t.Fatalf("Error compiling %v", err)
		} else {
			log.Printf("Build complete!")
		}

		cmd := exec.Command(path.Join(RootDir, "main"), "-addr", fmt.Sprintf(":%d", pt), "-api-addr", fmt.Sprintf(":%d", apt), "-v", "2")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Fatal("Cannot run martian")
		}

		ech := make(chan error)
		go func() {
			_, err := cmd.Process.Wait()
			ech <- err
		}()

		select {
		case <-mch:
			err := cmd.Process.Kill()
			if err != nil {
				log.Fatalf("Failed to kill martian %v", err)
			} else {
				log.Println("Killed martian")
			}
		case err := <-ech:
			log.Fatalf("Martian exited unexpectedly %v", err)
		}
		mch <- 1
	}(8889, 8890)
	defer func() {
		mch <- 1
		<-mch
	}()

	// Bring up a backend server, that counts from 0, and wait till it's ready
	var cs CounterServer = 0
	bs := &http.Server{Addr: fmt.Sprintf(":%d", 8891), Handler: &cs}
	go bs.ListenAndServe()

	v, err := -1, errors.New("")
	for err != nil {
		v, err = fetchValueByRequest("http://localhost:8891", "")
	}
	if v != 0 {
		t.Error("Expecting 0 as the first response")
	}

	// Turn on recording
	pu := "http://localhost:8889"

	PostJsonConfigToMartian(pu, `{"cache.Modifier": {  "scope": ["response"],  "mode": "cache"}}`)

	// Send a request, should get 1. It is then also recorded.
	if v, err = fetchValueByRequest("http://localhost:8891", pu); err != nil {
		t.Error(err)
	} else if v != 1 {
		t.Error("Expecting to get 1, but got %d", v)
	}

	// Turn on replay
	PostJsonConfigToMartian(pu, `{"cache.Modifier": {  "scope": ["response"],  "mode": "replay"}}`)

	// Send two requests, the backend server gets no requests (when directly get, we get 1)
	if v, err = fetchValueByRequest("http://localhost:8891", pu); err != nil {
		t.Error(err)
	} else if v != 1 {
		t.Error("Expecting to get 1, but got", v)
	}

	if v, err = fetchValueByRequest("http://localhost:8891", pu); err != nil {
		t.Error(err)
	} else if v != 1 {
		t.Error("Expecting to get 1, but got", v)
	}

	if cs != 2 {
		t.Error("Expecting the server to have received 2 requests, but have actually received", cs)
	}
}
