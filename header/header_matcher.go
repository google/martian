package header

import (
	"net/http"

	"github.com/google/martian/proxyutil"
)

type Matcher struct {
	name, value string
}

func NewMatcher(name, value string) *Matcher {
	return &Matcher{
		name:  name,
		value: value,
	}
}

func (m *Matcher) MatchRequest(req *http.Request) bool {
	h := proxyutil.RequestHeader(req)

	vs, ok := h.All(m.name)
	if !ok {
		return false
	}

	for _, v := range vs {
		if v == m.value {
			return true
		}
	}

	return false
}

func (m *Matcher) MatchResponse(res *http.Response) bool {
	h := proxyutil.ResponseHeader(res)

	vs, ok := h.All(m.name)
	if !ok {
		return false
	}

	for _, v := range vs {
		if v == m.value {
			return true
		}
	}

	return false
}
