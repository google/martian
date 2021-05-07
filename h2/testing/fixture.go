// Package testing contains a test fixture for working with gRPC over HTTP/2.
package testing

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/h2"
	"github.com/google/martian/v3/mitm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	tspb "github.com/google/martian/v3/h2/testservice"
)

var (
	// proxyPort is a global variable that stores the listener used by the proxy. This value is
	// shared globally because golang http transport code caches the environment variable values, in
	// particular HTTPS_PROXY.
	proxyPort int
)

// Fixture encapsulates the TestService gRPC server, a proxy and a gRPC client.
type Fixture struct {
	// TestServiceClient is a client pointing at the service and redirected through the proxy.
	tspb.TestServiceClient

	wg     sync.WaitGroup
	server *grpc.Server
	// serverErr is any error returned by invoking `Serve` on the gRPC server.
	serverErr error

	proxyListener net.Listener
	proxy         *martian.Proxy

	conn *grpc.ClientConn
}

// New creates a new instance of the Fixture. It is not possible for there to be more than one
// instance concurrently because clients decide whether to use the proxy based on the global
// HTTPS_PROXY environment variable.
func New(spf []h2.StreamProcessorFactory) (*Fixture, error) {
	f := &Fixture{}

	// Starts the gRPC server.
	f.server = grpc.NewServer(grpc.Creds(credentials.NewServerTLSFromCert(Localhost)))
	tspb.RegisterTestServiceServer(f.server, &Server{})

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("creating listener for gRPC service: %w", err)
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.serverErr = f.server.Serve(lis)
	}()

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("getting hostname: %w", err)
	}

	// Creates a listener for the proxy, obtaining a new port if needed.
	if proxyPort == 0 {
		// Attempts a query to port server first, falling back if it is unavailable. Ports that are
		// provided by listening on ":0" can be recyled by the OS leading to flakiness in certain
		// environments since we need the same port to be available across multiple instances of the
		// test fixture.
		proxyPort = queryPortServer()
		if proxyPort == 0 {
			var err error
			f.proxyListener, err = net.Listen("tcp", ":0")
			if err != nil {
				return nil, fmt.Errorf("creating listener for proxy; %w", err)
			}
			proxyPort = f.proxyListener.Addr().(*net.TCPAddr).Port
		}
		proxyTarget := hostname + ":" + strconv.Itoa(proxyPort)
		// Sets the HTTPS_PROXY environment variable so that http requests will go through the proxy.
		os.Setenv("HTTPS_PROXY", fmt.Sprintf("http://%s", proxyTarget))
		fmt.Printf("proxy at %s\n", proxyTarget)
	}
	if f.proxyListener == nil {
		var err error
		f.proxyListener, err = net.Listen("tcp", fmt.Sprintf(":%d", proxyPort))
		if err != nil {
			return nil, fmt.Errorf("creating listener for proxy; %w", err)
		}
	}

	// Starts the proxy.
	f.proxy, err = newProxy(spf)
	if err != nil {
		return nil, fmt.Errorf("creating proxy: %w", err)
	}
	go func() {
		f.proxy.Serve(f.proxyListener)
	}()

	port := lis.Addr().(*net.TCPAddr).Port
	target := hostname + ":" + strconv.Itoa(port)

	fmt.Printf("server at %s\n", target)

	// Connects a gRPC client with the service via the proxy.
	f.conn, err = grpc.Dial(target, grpc.WithTransportCredentials(ClientTLS))
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %w", target, err)
	}
	f.TestServiceClient = tspb.NewTestServiceClient(f.conn)

	return f, nil
}

// Close cleans up the servers and connections.
func (f *Fixture) Close() error {
	f.conn.Close()
	f.server.Stop()
	f.proxy.Close()
	f.wg.Wait()

	if err := f.proxyListener.Close(); err != nil {
		return fmt.Errorf("closing proxy listener: %w", err)
	}
	return f.serverErr
}

func newProxy(spf []h2.StreamProcessorFactory) (*martian.Proxy, error) {
	p := martian.NewProxy()
	mc, err := mitm.NewConfig(CA, CAKey)
	if err != nil {
		return nil, fmt.Errorf("creating mitm config: %w", err)
	}
	mc.SetValidity(time.Hour)
	mc.SetOrganization("Martian Proxy")
	mc.SetH2Config(&h2.Config{
		AllowedHostsFilter:       func(_ string) bool { return true },
		RootCAs:                  RootCAs,
		StreamProcessorFactories: spf,
		EnableDebugLogs:          true,
	})

	p.SetMITM(mc)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: RootCAs,
		},
	}
	p.SetRoundTripper(tr)

	return p, nil
}

func queryPortServer() int {
	// portpicker isn't available in third_party.
	if portServer := os.Getenv("PORTSERVER_ADDRESS"); portServer != "" {
		c, err := net.Dial("unix", portServer)
		if err != nil {
			// failed connection to portServer; this is normal in many circumstances.
			return 0
		}
		defer c.Close()
		if _, err := fmt.Fprintf(c, "%d\n", os.Getpid()); err != nil {
			return 0
		}
		buf, err := ioutil.ReadAll(c)
		if err != nil || len(buf) == 0 {
			return 0
		}
		buf = buf[:len(buf)-1] // remove newline char
		port, err := strconv.Atoi(string(buf))
		if err != nil {
			return 0
		}
		fmt.Printf("got port %d\n", port)
		return port
	}
	return 0
}
