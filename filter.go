package martian

type Filter struct {
	reqcond RequestCondition
	rescond ResponseCondition
	reqmod  RequestModifier
	resmod  ResponseModifier
}

func (f *Filter) SetRequestCondition(reqcond RequestCondition, reqmod RequestModifier, elsemod RequestModifier) {
	f.reqcond = reqcond
	f.reqmod = reqmod
}

func (f *Filter) SetResponseCondition(rescond ResponseCondition, resmod ResponseModifer) {
	f.rescond = rescond
	f.resmod = resmod
}

func (f *filter) ModifyRequest(req *http.Request) error {
	if !f.reqcond.MatchRequest(req) {
		return nil
	}
	return f.reqmod.ModifyRequest(req)
}

func (f *filter) ModifyResponse(res *http.Response) error {
	if !f.rescond.MatchRequest(req) {
		return nil
	}
	return f.resmod.ModifyResponse(res)
}
