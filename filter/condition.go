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
