package querystring

import "net/http"

// Matcher is a conditonal evalutor of query string parameters
// to be used in structs that take conditions.
type Matcher struct {
	name, value string
}

// NewMatcher builds a new querystring matcher
func NewMatcher(name, value string) *Matcher {
	return &Matcher{name: name, value: value}
}

// MatchRequest evaluates a request and returns whether or not
// the request contains a querystring param that matches the provided name
// and value.
func (m *Matcher) MatchRequest(req *http.Request) (bool, error) {
	for n, vs := range req.URL.Query() {
		if m.name == n {
			if m.value == "" {
				return true, nil
			}

			for _, v := range vs {
				if m.value == v {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// MatchResponse evaluates a response and returns whether or not
// the request that resulted in that response contains a querystring param that matches the provided name
// and value.
func (m *Matcher) MatchResponse(res *http.Response) (bool, error) {
	return m.MatchRequest(res.Request)
}
