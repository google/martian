package martianurl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/google/martian/parse"
)

func init() {
	parse.Register("urlcount.Verifier", urlCountVerifierFromJSON)
}

type urlCountVerifier struct {
	urlPattern *regexp.Regexp
	params     map[string]*regexp.Regexp
	subParams  map[string]map[string]*regexp.Regexp
	count      int

	mu      sync.Mutex
	counted int
}

type urlCountVerifierJSON struct {
	Count      int                    `json:"count"`
	URLPattern string                 `json:"urlPattern"`
	Params     map[string]interface{} `json:"params"`
	Scope      []parse.ModifierType   `json:"scope"`
}

func NewURLCountVerifier(urlpattern string, count int) (*urlCountVerifier, error) {
	r, err := regexp.Compile("^" + urlpattern + "$")
	if err != nil {
		return nil, fmt.Errorf("urlcount: error generating regexp from string %v", urlpattern)
	}
	return &urlCountVerifier{
		urlPattern: r,
		count:      count,
		params:     make(map[string]*regexp.Regexp),
		subParams:  make(map[string]map[string]*regexp.Regexp),
	}, nil
}

// ModifyRequest checks if a request's url matches urlPattern. If it does, counted is incremented by
// 1. Note that this does not write anything to err; determining if there has been an error or not
// is handled in VerifyRequests().
func (v *urlCountVerifier) ModifyRequest(req *http.Request) error {
	if !v.urlPattern.MatchString(req.URL.String()) {
		// Not an error, just not a match.
		return nil
	}

	query := req.URL.Query()

	isMatch := true
	for k, regex := range v.params {
		if _, ok := query[k]; !ok {
			isMatch = false
			break
		}
		if !regex.MatchString(query.Get(k)) {
			isMatch = false
			break
		}
	}

outer:
	for k, m := range v.subParams {
		rss := query.Get(k)
		if rss == "" {
			isMatch = false
			break
		}
		rsm, err := url.ParseQuery(rss)
		if err != nil {
			isMatch = false
			break
		}
		for vsk, vsv := range m {
			if !vsv.MatchString(rsm.Get(vsk)) {
				isMatch = false
				break outer
			}
		}
	}
	if isMatch {
		v.mu.Lock()
		v.counted++
		v.mu.Unlock()
	}
	return nil
}

// VerifyRequests returns an error if counted is not equal to count.
func (v *urlCountVerifier) VerifyRequests() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.count != v.counted {
		if len(v.subParams) > 0 {
			return fmt.Errorf("urlcount verification error: expected url pattern %v with params %+v and subparams %+v to be pinged %d times; was pinged %d times", v.urlPattern, v.params, v.subParams, v.count, v.counted)
		}
		return fmt.Errorf("urlcount verification error: expected url pattern %v with params %+v to be pinged %d times; was pinged %d times", v.urlPattern, v.params, v.count, v.counted)
	}
	return nil
}

// ResetRequestVerifications resets counted.
func (v *urlCountVerifier) ResetRequestVerifications() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.counted = 0
}

// urlCountVerifierFromJSON builds a urlcount.Verifier from JSON.
//
// Example JSON:
// {
//   "urlcount.Verifier": {
//     "scope": ["request"],
//     "urlPattern": "example.com/.*",
//     "count": 1,
//     "params": {
//	     "label": "videoplayfailed.+",
//       "label2": {
//         "subkey": "subvalue.+",
//         "subkey2": "subvalue2"
//       }
//     }
//   }
// }
func urlCountVerifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &urlCountVerifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	v, err := NewURLCountVerifier(msg.URLPattern, msg.Count)
	if err != nil {
		return nil, fmt.Errorf("error generating regexp from string %v", msg.URLPattern)
	}

	for key, val := range msg.Params {
		switch p := val.(type) {
		case string:
			r, err := regexp.Compile("^" + p + "$")
			if err != nil {
				return nil, fmt.Errorf("error generating regexp from string: %v for key %q", p, key)
			}
			v.params[key] = r
		case map[string]interface{}:
			submap := make(map[string]*regexp.Regexp)
			for subkey, subval := range p {
				strsubval, ok := subval.(string)
				if !ok {
					return nil, fmt.Errorf("could not convert subparameter to string: %v", subval)
				}
				r, err := regexp.Compile("^" + strsubval + "$")
				if err != nil {
					return nil, fmt.Errorf("error generating regexp from string: %v for key %q:%q", p, key, subkey)
				}
				submap[subkey] = r
			}
			v.subParams[key] = submap
		default:
			return nil, fmt.Errorf("invalid params json: %v", val)
		}
	}
	return parse.NewResult(v, msg.Scope)
}
