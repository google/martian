package filter

import (
	"net/http"

	"github.com/google/martian"
)

var noop = martian.Noop("Filter")

type Filter struct {
	reqcond RequestCondition
	rescond ResponseCondition

	treqmod martian.RequestModifier
	tresmod martian.ResponseModifier
	freqmod martian.RequestModifier
	fresmod martian.ResponseModifier
}

func New() *Filter {
	return &Filter{
		treqmod: noop,
		tresmod: noop,
		fresmod: noop,
		freqmod: noop,
	}
}

func (f *Filter) SetRequestCondition(reqcond RequestCondition) {
	f.reqcond = reqcond
}

func (f *Filter) SetResponseCondition(rescond ResponseCondition) {
	f.rescond = rescond
}

func (f *Filter) RequestWhenTrue(mod martian.RequestModifier) {
	f.treqmod = mod
}

func (f *Filter) ResponseWhenTrue(mod martian.ResponseModifier) {
	f.tresmod = mod
}

func (f *Filter) RequestWhenFalse(mod martian.RequestModifier) {
	f.freqmod = mod
}

func (f *Filter) ResponseWhenFalse(mod martian.ResponseModifier) {
	f.fresmod = mod
}

func (f *Filter) ModifyRequest(req *http.Request) error {
	if f.reqcond.MatchRequest(req) {
		return f.treqmod.ModifyRequest(req)
	}

	return f.freqmod.ModifyRequest(req)
}

func (f *Filter) ModifyResponse(res *http.Response) error {
	if f.rescond.MatchResponse(res) {
		return f.tresmod.ModifyResponse(res)
	}

	return f.fresmod.ModifyResponse(res)
}
