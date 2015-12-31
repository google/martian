package verify

import "github.com/google/martian"

const key = "verify.Context"

// Context contains an ordered set of error builders that may or may not
// produce verification errors depending on the conditions at retrieval time.
type context struct {
	set map[*ErrorBuilder]struct{}
	ebs []*ErrorBuilder
}

// Verify adds the error builder to the context.
func Verify(ctx *martian.Context, eb *ErrorBuilder) {
	v, ok := ctx.Get(key)
	if !ok {
		vctx := newContext()
		vctx.add(eb)

		ctx.Set(key, vctx)
	}

	v.(*context).add(eb)
}

// FromContext retrieves the error builders from the context.
func FromContext(ctx *martian.Context) []*ErrorBuilder {
	v, ok := ctx.Get(key)
	if !ok {
		return nil
	}

	return v.(*context).ebs
}

func newContext() *context {
	return &context{
		set: make(map[*ErrorBuilder]struct{}),
	}
}

func (ctx *context) add(eb *ErrorBuilder) {
	if _, ok := ctx.set[eb]; ok {
		return
	}

	ctx.set[eb] = struct{}{}
	ctx.ebs = append(ctx.ebs, eb)
}
