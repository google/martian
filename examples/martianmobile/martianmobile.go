// Package martianmobile is a small subset of the Martian API intended to be
// built with gomobile for Android and iOS support.
package martianmobile

import (
	"log"
	"net"
	"net/http"

	_ "github.com/google/martian/body"
	_ "github.com/google/martian/cookie"
	_ "github.com/google/martian/header"
	_ "github.com/google/martian/method"
	_ "github.com/google/martian/pingback"
	_ "github.com/google/martian/querystring"
	_ "github.com/google/martian/status"

	"github.com/google/martian"
	"github.com/google/martian/fifo"
	"github.com/google/martian/martianhttp"
	"github.com/google/martian/verify"
)

// Proxy is a wrapper for the initialized Martian proxy.
type Proxy struct {
        proxy    *martian.Proxy
	listener net.Listener
}

// Start runs a martian.Proxy on addr.
func Start(addr string) (*Proxy, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	p := martian.NewProxy(nil)
	m := martianhttp.NewModifier()
	fg := fifo.NewGroup()

	fg.AddRequestModifier(m)
	fg.AddResponseModifier(m)

	p.SetRequestModifier(fg)
	p.SetResponseModifier(fg)

	// Proxy specific handlers.
	// These handlers take precendence over proxy traffic and will not be
	// intercepted.

	// Update modifiers.
	http.Handle("/martian/modifiers", m)

	// Verify assertions.
	vh := verify.NewHandler()
	vh.SetRequestVerifier(m)
	vh.SetResponseVerifier(m)
	http.Handle("/martian/verify", vh)

	// Reset verifications.
	rh := verify.NewResetHandler()
	rh.SetRequestVerifier(m)
	rh.SetResponseVerifier(m)
	http.Handle("/martian/verify/reset", rh)

        http.Handle("/", p)

        log.Printf("Martian proxy starting\n")
        go http.Serve(l, nil)

        return &Proxy{
            proxy: p,
            listener: l,
        }, nil
}

// Shutdown closes the martian.Proxy.
func (p *Proxy) Shutdown() {
	p.listener.Close()
        log.Printf("Martian proxy shut down\n")
}
