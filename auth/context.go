package auth

import "github.com/google/martian/session"

const key = "auth.Context"

type Context struct {
	ctx *session.Context
	err error
}

func FromContext(ctx *session.Context) *Context {
	v, ok := ctx.Get(key)
	if !ok {
		actx := &Context{
			ctx: ctx,
		}
		ctx.Set(key, actx)
		return actx
	}

	return v.(*Context)
}

func (ctx *Context) ID() string {
	return ctx.ctx.SessionID()
}

func (ctx *Context) SetID(id string) {
	ctx.err = nil

	if id == "" {
		return
	}

	ctx.ctx.SetSessionID(id)
}

func (ctx *Context) SetError(err error) {
	ctx.ctx.SetSessionID("")
	ctx.err = err
}

func (ctx *Context) Error() error {
	return ctx.err
}
