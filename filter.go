package martian

import (
	"net/http"
)

type Filter struct {
	reqcond RequestCondition
	rescond ResponseCondition
	reqmod  RequestModifier
	resmod  ResponseModifier
}

func (f *Filter) SetRequestCondition(reqcond RequestCondition, reqmod RequestModifier) {
	f.reqcond = reqcond
	f.reqmod = reqmod
}

func (f *Filter) SetResponseCondition(rescond ResponseCondition, resmod ResponseModifier) {
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
