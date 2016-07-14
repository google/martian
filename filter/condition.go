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

// type Condition struct {
// 	reqcon RequestCondition
// }
//
// func (c *Condition) RequestCondition() RequestCondition {
// 	return c.reqcon
// }
//
// func (c *Condition) ResponseCondition() ResponseCondition {
// 	return c.rescon
// }
//
// func (c *Condition) MatchRequest(req *http.Request) bool {
// 	return c.MatchRequest(req)
// }
//
// func (c *Condition) MatchResponse(res *http.Response) bool {
// 	return c.MatchResponse(res)
// }
//
// type NotCondition struct {
// 	reqcon RequestCondition
// 	rescon ResponseCondition
// }
//
// func (nc *NotCondition) RequestCondition() RequestCondition {
// 	return nc.reqcon
// }
//
// func (nc *NotCondition) ResponseCondition() ResponseCondition {
// 	return nc.rescon
// }
//
// func (nc *NotCondition) MatchRequest(req *http.Request) bool {
// 	return !nc.MatchRequest(req)
// }
//
// func (nc *NotCondition) MatchResponse(res *http.Response) bool {
// 	return !nc.MatchResponse(res)
// }
