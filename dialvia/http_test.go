// Copyright 2023 Sauce Labs, Inc. All rights reserved.

package dialvia

import (
	"bufio"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/martian/v3/proxyutil"
	"golang.org/x/net/context"
)

func TestHTTPProxyDialerDialContext(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	d := HTTPProxy(
		(&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		&url.URL{Scheme: "http", Host: l.Addr().String()},
	)

	ctx := context.Background()

	t.Run("status 200", func(t *testing.T) {
		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				pbr := bufio.NewReader(conn)
				req, err := http.ReadRequest(pbr)
				if err != nil {
					return err
				}
				return proxyutil.NewResponse(200, nil, req).Write(conn)
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err != nil {
			t.Fatal(err)
		}
		if conn == nil {
			t.Fatal("conn is nil")
		}

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("connect status 404", func(t *testing.T) {
		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				pbr := bufio.NewReader(conn)
				req, err := http.ReadRequest(pbr)
				if err != nil {
					return err
				}
				return proxyutil.NewResponse(404, nil, req).Write(conn)
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err == nil {
			t.Fatal("err is nil")
		}
		t.Log(err)
		if conn != nil {
			t.Fatal("conn is not nil")
		}

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("conn closed", func(t *testing.T) {
		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				conn.Close()
				return nil
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err == nil {
			t.Fatal("err is nil")
		}
		t.Log(err)
		if conn != nil {
			t.Fatal("conn is not nil")
		}

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		done := make(chan struct{})
		go func() {
			serveOne(l, func(conn net.Conn) error {
				cancel()
				<-done
				return nil
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err == nil {
			t.Fatal("err is nil")
		}
		t.Log(err)
		if conn != nil {
			t.Fatal("conn is not nil")
		}

		done <- struct{}{}
	})
}

func serveOne(l net.Listener, h func(conn net.Conn) error) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	return h(conn)
}
