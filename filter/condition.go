package filter

import (
	"net/http"
)

type ResponseCondition interface {
	MatchResponse(*http.Response) bool
}

type RequestCondition interface {
	MatchRequest(*http.Request) bool
}

type Condition struct {
	Request  RequestCondition
	Response ResponseCondition
}

func (c *Condition) MatchRequest(req *http.Request) bool {
	return c.MatchRequest(req)
}

func (c *Condition) MatchResponse(res *http.Response) bool {
	return c.MatchResponse(res)
}

type NotCondition struct {
	Request  RequestCondition
	Response ResponseCondition
}

func (nc *NotCondition) MatchRequest(req *http.Request) bool {
	return !nc.MatchRequest(req)
}

func (nc *NotCondition) MatchResponse(res *http.Response) bool {
	return !nc.MatchResponse(res)
}

func Invert(c Condition) *NotCondition {
	return &NotCondition{
		Request:  c.Request,
		Response: c.Response,
	}
}
