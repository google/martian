package verify

import (
	"fmt"

	"github.com/google/martian"
)

const key = "verify.Context"

// Context contains a list of functions that may or may not produce
// verification errors related to a given request/response pair.
type Context struct {
	errs []Error
}

// ForContext adds the verification error to the context.
func ForContext(ctx *martian.Context, err Error) error {
	v, ok := ctx.Get(key)
	if !ok {
		return fmt.Errorf("verify: missing error context")
	}

	vctx := v.(*Context)
	vctx.errs = append(vctx.errs, err)

	return nil
}

// FromContext retrieves the verification errors from the context.
func FromContext(ctx *martian.Context) []Error {
	v, ok := ctx.Get(key)
	if !ok {
		return nil
	}

	return v.(*Context).errs
}
