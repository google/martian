package verify

import (
	"net/http"
	"sync"

	"github.com/google/martian"
)

// Verification collects all errors from the contained request and response
// verifiers.
type Verification struct {
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier

	mu   sync.RWMutex
	errs []Error
}

var noop = martian.Noop("verify.Group")

// NewVerification returns a new verification suite.
func NewVerification() *Verification {
	return &Verification{
		reqmod: noop,
		resmod: noop,
		errs:   make([]Error, 0),
	}
}

// SetRequestModifier sets the request modifier.
func (v *Verification) SetRequestModifier(reqmod martian.RequestModifier) {
	if reqmod == nil {
		v.reqmod = noop
	}

	v.reqmod = reqmod
}

// SetResponseModifier sets the response modifier.
func (v *Verification) SetResponseModifier(resmod martian.ResponseModifier) {
	if resmod == nil {
		v.resmod = noop
	}

	v.resmod = resmod
}

// ModifyRequest sets up the error context for later verifiers.
func (v *Verification) ModifyRequest(req *http.Request) error {
	NewContext(martian.NewContext(req))

	return nil
}

// ModifyResponse collects the verification errors from the error context.
func (v *Verification) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	errs := FromContext(ctx)

	v.mu.Lock()
	defer v.mu.Unlock()

	v.errs = append(v.errs, errs...)

	return nil
}

// Errors returns the list of verification errors.
func (v *Verification) Errors() []Error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.errs
}

// Reset resets the verification errors and calls Reset() for any errors that
// implement verify.Resetter.
func (v *Verification) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, err := range v.errs {
		if r, ok := err.(Resetter); ok {
			r.Reset()
		}
	}

	v.errs = make([]Error, 0)
}
