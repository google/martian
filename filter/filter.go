package filter

import (
	"net/http"

	"github.com/google/martian"
)

var noop = martian.Noop("Filter")

type Filter struct {
	reqcond   RequestCondition
	rescond   ResponseCondition
	posreqmod martian.RequestModifier
	posresmod martian.ResponseModifier
	negreqmod martian.RequestModifier
	negresmod martian.ResponseModifier
}

func New() *Filter {
	return &Filter{
		posreqmod: noop,
		posresmod: noop,
		negresmod: noop,
		negreqmod: noop,
	}
}

func (f *Filter) SetRequestCondition(reqcond RequestCondition) {
	f.reqcond = reqcond
}

func (f *Filter) SetResponseCondition(rescond ResponseCondition) {
	f.rescond = rescond
}

func (f *Filter) SetRequestModifiers(posmod martian.RequestModifier, negmod martian.RequestModifier) {
	if posmod == nil {
		posmod = noop
	}
	if negmod == nil {
		negmod = noop
	}
	f.posreqmod = posmod
	f.negreqmod = negmod
}

func (f *Filter) SetResponseModifiers(posmod martian.ResponseModifier, negmod martian.ResponseModifier) {
	if posmod == nil {
		posmod = noop
	}
	if negmod == nil {
		negmod = noop
	}
	f.posresmod = posmod
	f.negresmod = negmod
}

func (f *Filter) ModifyRequest(req *http.Request) error {
	if f.reqcond.MatchRequest(req) {
		return f.posreqmod.ModifyRequest(req)
	}

	return f.negreqmod.ModifyRequest(req)
}

func (f *Filter) ModifyResponse(res *http.Response) error {
	if f.posresmod != nil && f.rescond.MatchResponse(res) {
		return f.posresmod.ModifyResponse(res)
	}

	if f.negresmod != nil && !f.rescond.MatchResponse(res) {
		return f.negresmod.ModifyResponse(res)
	}

	return nil
}
