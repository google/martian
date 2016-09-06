// Package api contains a forwarder to route system HTTP requests to the API server.
package api

import (
	"net/http"

	"github.com/google/martian"
)

// Forwarder is a request modifier that routes the request to the API server and
// marks the request for skipped logging.
type Forwarder struct {
	host string
}

// NewForwarder returns a Forwarder that rewrites requests to host.
func NewForwarder(host string) *Forwarder {
	return &Forwarder{
		host: host,
	}
}

// ModifyRequest changes the request host to f.Host, downgrades the scheme to http
// and marks the request context for skipped logging.
func (f *Forwarder) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	ctx.SkipLogging()

	req.URL.Scheme = "http"
	req.URL.Host = f.host

	return nil
}
