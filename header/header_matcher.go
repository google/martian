package header

import (
	"net/http"

	"github.com/google/martian/proxyutil"
)

// Matcher is a conditonal evalutor of request or
// response headers to be used in structs that take conditions.
type Matcher struct {
	name, value string
}

// NewMatcher builds a new header matcher.
func NewMatcher(name, value string) *Matcher {
	return &Matcher{
		name:  name,
		value: value,
	}
}

// MatchRequest evaluates a request and returns whether or not
// the request contains a header that matches the provided name
// and value.
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

// MatchResponse evaluates a response and returns whether or not
// the response contains a header that matches the provided name
// and value.
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
