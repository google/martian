package querystring

import "net/http"

type Matcher struct {
	name, value string
}

// NewMatcher builds a new querystring matcher
func NewMatcher(name, value string) *Matcher {
	return &Matcher{name: name, value: value}
}

func (m *Matcher) MatchRequest(req *http.Request) bool {
	for n, vs := range req.URL.Query() {
		if m.name == n {
			if m.value == "" {
				return true
			}

			for _, v := range vs {
				if m.value == v {
					return true
				}
			}
		}
	}

	return false
}

func (m *Matcher) MatchResponse(res *http.Response) bool {
	return m.MatchRequest(res.Request)
}
