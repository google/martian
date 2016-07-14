package filter

import (
	"net/http"

	"github.com/google/martian"
)

type Filter struct {
	reqcond RequestCondition
	rescond ResponseCondition
	reqmod  martian.RequestModifier
	resmod  martian.ResponseModifier
}

func (f *Filter) SetRequestCondition(reqcond RequestCondition, reqmod martian.RequestModifier) {
	f.reqcond = reqcond
	f.reqmod = reqmod
}

func (f *Filter) SetResponseCondition(rescond ResponseCondition, resmod martian.ResponseModifier) {
	f.rescond = rescond
	f.resmod = resmod
}

func (f *Filter) ModifyRequest(req *http.Request) error {
	if !f.reqcond.MatchRequest(req) {
		return nil
	}
	return f.reqmod.ModifyRequest(req)
}

func (f *Filter) ModifyResponse(res *http.Response) error {
	if !f.rescond.MatchResponse(res) {
		return nil
	}
	return f.resmod.ModifyResponse(res)
}
