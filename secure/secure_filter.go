// Package secure provides filters that allow for the conditional execution
// of modifiers based on whether or not the connection is secured with TLS.
package secure

import (
	"net/http"

	"github.com/google/martian"
)

var noop = martian.Noop("secure.Filter")

// Filter provides conditional execution of reqmod on the basis of
// whether or not the connection is secure.
type Filter struct {
	secure bool
	reqmod martian.RequestModifier
}

// NewUnsecureFilter returns a filter that executes reqmod when the
// connection is secured by TLS
func NewSecureFilter() *Filter {
	return &Filter{
		secure: true,
		reqmod: noop,
	}
}

// NewUnsecureFilter returns a filter that executes reqmod when the
// connection is not secured by TLS.
func NewUnsecureFilter() *Filter {
	return &Filter{
		secure: false,
		reqmod: noop,
	}
}

// SetRequestModifier sets the request modifier of filter.
func (f *Filter) SetRequestModifier(reqmod martian.RequestModifier) {
	if reqmod == nil {
		f.reqmod = noop
		return
	}

	f.reqmod = reqmod
}

// ModifyRequest runs ModifyRequest on reqmod when the session is unsecured
// for an UnsecureFilter and the session is secure for a SecureFilter.
func (f *Filter) ModifyRequest(req *http.Request) error {
	ctx := martian.Context(req)

	if f.secure == ctx.GetSession().IsSecure() {
		return f.reqmod.ModifyRequest(req)
	}

	return nil
}
