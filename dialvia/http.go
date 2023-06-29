// Copyright 2023 Sauce Labs, Inc. All rights reserved.

package dialvia

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type HTTPProxyDialer struct {
	dial      ContextDialerFunc
	proxyURL  *url.URL
	tlsConfig *tls.Config
}

func HTTPProxy(dial ContextDialerFunc, proxyURL *url.URL) *HTTPProxyDialer {
	if dial == nil {
		panic("dial is required")
	}
	if proxyURL == nil {
		panic("proxy URL is required")
	}
	if proxyURL.Scheme != "http" {
		panic("proxy URL scheme must be http")
	}

	return &HTTPProxyDialer{
		dial:     dial,
		proxyURL: proxyURL,
	}
}

func HTTPSProxy(dial ContextDialerFunc, proxyURL *url.URL, tlsConfig *tls.Config) *HTTPProxyDialer {
	if dial == nil {
		panic("dial is required")
	}
	if proxyURL == nil {
		panic("proxy URL is required")
	}
	if proxyURL.Scheme != "https" {
		panic("proxy URL scheme must be https")
	}
	if tlsConfig == nil {
		panic("TLS config is required")
	}

	tlsConfig.ServerName = proxyURL.Hostname()
	tlsConfig.NextProtos = []string{"http/1.1"}

	return &HTTPProxyDialer{
		dial:      dial,
		proxyURL:  proxyURL,
		tlsConfig: tlsConfig,
	}
}

func (d *HTTPProxyDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	res, conn, err := d.DialContextR(ctx, network, addr)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			b = []byte(fmt.Sprintf("error dumping response: %s", err))
		}

		conn.Close()
		return nil, fmt.Errorf("proxy connection failed status=%d\n\n%s", res.StatusCode, string(b))
	}

	return conn, nil
}

func (d *HTTPProxyDialer) DialContextR(ctx context.Context, network, addr string) (*http.Response, net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, nil, fmt.Errorf("unsupported network: %s", network)
	}

	conn, err := d.dial(ctx, "tcp", d.proxyURL.Host)
	if err != nil {
		return nil, nil, err
	}
	if d.proxyURL.Scheme == "https" {
		conn = tls.Client(conn, d.tlsConfig)
	}

	pbw := bufio.NewWriterSize(conn, 1024)
	pbr := bufio.NewReaderSize(conn, 1024)

	connReq := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: addr},
		Host:   addr,
	}
	if d.proxyURL.User != nil {
		connReq.Header = make(http.Header, 1)
		connReq.Header.Add("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(d.proxyURL.User.String())))
	}

	if err := connReq.Write(pbw); err != nil {
		conn.Close()
		return nil, nil, err
	}
	if err := pbw.Flush(); err != nil {
		conn.Close()
		return nil, nil, err
	}

	res, err := http.ReadResponse(pbr, connReq)
	return res, conn, err
}
