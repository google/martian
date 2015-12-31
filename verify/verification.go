package verify

import (
	"net/http"
	"sync"

	"github.com/google/martian"
)

// Verification keeps an ordered set of all errors from the contained request
// and response verifiers.
type Verification struct {
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier

	mu  sync.RWMutex
	set map[*ErrorBuilder]struct{}
	ebs []*ErrorBuilder
}

var noop = martian.Noop("verify.Group")

// NewVerification returns a new verification suite.
func NewVerification() *Verification {
	return &Verification{
		reqmod: noop,
		resmod: noop,
		set:    make(map[*ErrorBuilder]struct{}),
		ebs:    make([]*ErrorBuilder, 0),
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

// ModifyRequest runs request modifiers and collects any verification errors.
func (v *Verification) ModifyRequest(req *http.Request) error {
	err := v.reqmod.ModifyRequest(req)
	ctx := martian.NewContext(req)

	v.saveErrors(ctx)

	return err
}

// ModifyResponse runs response modifiers and collects any verification errors.
func (v *Verification) ModifyResponse(res *http.Response) error {
	err := v.resmod.ModifyResponse(res)
	ctx := martian.NewContext(res.Request)

	v.saveErrors(ctx)

	return err
}

// saveErrors loops through the error builders from the context and saves any
// new error builders. Existing error builders will be skipped.
func (v *Verification) saveErrors(ctx *martian.Context) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, eb := range FromContext(ctx) {
		if _, ok := v.set[eb]; ok {
			continue
		}

		v.ebs = append(v.ebs, eb)
	}
}

// Errors returns the list of verification errors.
func (v *Verification) Errors() []Error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	verrs := make([]Error, len(v.ebs))

	for _, eb := range v.ebs {
		verr, ok := eb.Error()
		if !ok {
			continue
		}

		verrs = append(verrs, verr)
	}

	return verrs
}

// Reset resets the verification errors and any error builders.
func (v *Verification) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, eb := range v.ebs {
		eb.Reset()
	}

	v.set = make(map[*ErrorBuilder]struct{})
	v.ebs = make([]*ErrorBuilder, 0)
}
