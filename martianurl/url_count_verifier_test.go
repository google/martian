package martianurl

import (
	"net/http"
	"regexp"
	"sync"
	"testing"

	"github.com/google/martian/parse"
	"github.com/google/martian/verify"
)

func TestConcurrentRequests(t *testing.T) {
	v, err := NewURLCountVerifier("http://www.example.com/.*", 1000)
	if err != nil {
		t.Fatalf("urlcount.NewURLCountVerifier(): got %v, want no error", err)
	}

	var wg sync.WaitGroup
	var groupErr error
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		f := func() error {
			defer wg.Done()
			req, err := http.NewRequest("GET", "http://www.example.com/foo", nil)
			if err != nil {
				return err
			}
			if err := v.ModifyRequest(req); err != nil {
				t.Fatalf("ModifyRequest(): got %v, want no error", err)
			}
			return err
		}
		go f()
	}

	wg.Wait()
	if groupErr != nil {
		t.Errorf("unexpected error from StartSession: %v", groupErr)
	}
	if err := v.VerifyRequests(); err != nil {
		t.Fatalf("VerifyRequests(): got %v, want no error", err)
	}
}

func TestVerifyRequestPassesUrlPattern(t *testing.T) {
	v, err := NewURLCountVerifier("http://www.example.com/.*", 1)
	if err != nil {
		t.Fatalf("urlcount.NewURLCountVerifier(): got %v, want no error", err)
	}
	req, err := http.NewRequest("GET", "http://www.example.com/foo", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if err := v.VerifyRequests(); err != nil {
		t.Fatalf("VerifyRequests(): got %v, want no error", err)
	}
}

func TestVerifyRequestFailsCountedGreaterThanCount(t *testing.T) {
	v, err := NewURLCountVerifier("http://www.example.com/.*", 1)
	if err != nil {
		t.Fatalf("urlcount.NewURLCountVerifier(): got %v, want no error", err)
	}
	reqA, errA := http.NewRequest("GET", "http://www.example.com/foo", nil)
	if errA != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", errA)
	}
	reqB, errB := http.NewRequest("GET", "http://www.example.com/bar", nil)
	if errB != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", errB)
	}
	if err := v.ModifyRequest(reqA); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if err := v.ModifyRequest(reqB); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if err := v.VerifyRequests(); err == nil {
		t.Fatal("VerifyRequests(): expected error but didn't receive")
	}
}

func TestVerifyRequestFailureNotEnoughPings(t *testing.T) {
	v, err := NewURLCountVerifier("http://www.example.com/.*", 2)
	if err != nil {
		t.Fatalf("urlcount.NewURLCountVerifier(): got %v, want no error", err)
	}
	req, err := http.NewRequest("GET", "http://www.example.com/foo", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if err := v.VerifyRequests(); err == nil {
		t.Fatal("VerifyRequests(): got nil, want error")
	}
}

func TestVerifyRequestFailureTooMany(t *testing.T) {
	v, err := NewURLCountVerifier("http://www.example.com/.*", 0)
	if err != nil {
		t.Fatalf("urlcount.NewURLCountVerifier(): got %v, want no error", err)
	}
	req, err := http.NewRequest("GET", "http://www.example.com/foo", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}
	if err := v.VerifyRequests(); err == nil {
		t.Fatal("VerifyRequests(): got nil but expected error ")
	}
}

func TestVerifyRequestsFailureOnEmptyParamValue(t *testing.T) {
	// If the requested URL does not contain a given parameter, it should not
	// count as a match.
	v, err := NewURLCountVerifier("http://www.example.com/.*", 0)
	if err != nil {
		t.Fatalf("NewURLCountVerifier(): unexpected error: %v", err)
	}
	v.params = map[string]*regexp.Regexp{"foo": regexp.MustCompile(".*")}

	req, err := http.NewRequest("GET", "http://www.example.com/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): unexpected error: %v", err)
	}
	if err := v.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): unexpected error: %v", err)
	}
	if err := v.VerifyRequests(); err != nil {
		t.Fatalf("VerifyRequests(): got %v, want success", err)
	}
}

func TestVerifiers(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		reqURL string
		match  bool
	}{
		{
			name: "Simple URL only Match",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1
		     }
		  }`,
			reqURL: "http://www.example.com/foo",
			match:  true,
		},
		{
			name: "URL and Single Param Match",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1,
		       "params": {
		         "label": "vast_creativeview"
		       }
		     }
		  }`,
			reqURL: "http://www.example.com/foo?label=vast_creativeview",
			match:  true,
		},
		{
			name: "Url and Single Subparam Match",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1,
		       "params": {
		         "label": {
		         		"a": "foo"
		         }
		       }
		     }
		  }`,
			reqURL: "http://www.example.com/foo?label=a%3Dfoo",
			match:  true,
		},
		{
			name: "URL and Multiple Subparams Match",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1,
		       "params": {
		         "label": {
		         		"a": "foo",
		         		"b": "bar"
		         }
		       }
		     }
		  }`,
			reqURL: "http://www.example.com/foo?label=a%3Dfoo%3Bb%3Dbar",
			match:  true,
		},
		{
			name: "URL and Multiple Subparams in Multiple Params Match",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1,
		       "params": {
		         "label": {
		         		"a": "foo",
		         		"b": "bar"
		         },
		         "anotherlabel": {
		         		"c": "qux"
		         }
		       }
		     }
		  }`,
			reqURL: "http://www.example.com/foo?label=a%3Dfoo%3Bb%3Dbar&anotherlabel=c%3Dqux",
			match:  true,
		},
		{
			name: "URL and Mixed Parameters and Subparameters",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1,
		       "params": {
		         "label": {
		         		"a": "foo",
		         		"b": "bar"
		         },
		         "anotherlabel": "qux"
		       }
		     }
		  }`,
			reqURL: "http://www.example.com/foo?label=a%3Dfoo%3Bb%3Dbar&anotherlabel=qux",
			match:  true,
		},
		{
			name: "URL and Mixed Parameters and Subparameters - Parameter Mismatch",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1,
		       "params": {
		         "label": {
		         		"a": "foo",
		         		"b": "bar"
		         },
		         "anotherlabel": "dux"
		       }
		     }
		  }`,
			reqURL: "http://www.example.com/foo?label=a%3Dfoo%3Bb%3Dbar&anotherlabel=qux",
			match:  false,
		},
		{
			name: "URL and Mixed Parameters and Subparameters - Subparameter Mismatch",
			json: `{
		     "urlcount.Verifier": {
		       "scope": ["request"],
		       "urlPattern": "http://www.example.com/.*",
		       "count": 1,
		       "params": {
		         "label": {
		         		"a": "foo",
		         		"b": "jar"
		         },
		         "anotherlabel": "qux"
		       }
		     }
		  }`,
			reqURL: "http://www.example.com/foo?label=a%3Dfoo%3Bb%3Dbar&anotherlabel=qux",
			match:  false,
		},
	}

	for _, test := range tests {
		r, err := parse.FromJSON([]byte(test.json))
		if err != nil {
			t.Errorf("%s: parse.FromJSON(): got %v", test.name, err)
			continue
		}
		reqmod := r.RequestModifier()
		if reqmod == nil {
			t.Errorf("%s: reqmod: got nil, want not nil", test.name)
			continue
		}
		reqv, ok := reqmod.(verify.RequestVerifier)
		if !ok {
			t.Errorf("%s: reqmod.(verify.RequestVerifier): got !ok, want ok", test.name)
			continue
		}

		req, err := http.NewRequest("GET", test.reqURL, nil)
		if err != nil {
			t.Errorf("%s: http.NewRequest(): got %v, want no error", test.name, err)
			continue
		}
		if err := reqv.ModifyRequest(req); err != nil {
			t.Errorf("%s: ModifyRequest(): got %v, want no error", test.name, err)
			continue
		}
		err = reqv.VerifyRequests()
		if test.match && err != nil {
			t.Errorf("%s: VerifyRequests(): got error %v, want nil", test.name, err)
			continue
		}
		if !test.match && err == nil {
			t.Errorf("%s: VerifyRequest(): returned nil, but expected an error from not matching the expected verifier.", test.name)
		}

		reqv.ResetRequestVerifications()

		req, err = http.NewRequest("GET", test.reqURL, nil)
		if err != nil {
			t.Errorf("%s: http.NewRequest(): got %v, want no error", test.name, err)
			continue
		}
		if err := reqv.ModifyRequest(req); err != nil {
			t.Errorf("%s: ModifyRequest() after ResetRequestVerifications(): got %v, want no error", test.name, err)
			continue
		}
		err = reqv.VerifyRequests()
		if test.match && err != nil {
			t.Errorf("%s: VerifyRequests() after ResetRequestVerifications(): got error %v, want nil", test.name, err)
			continue
		}
		if !test.match && err == nil {
			t.Errorf("%s: VerifyRequest() after ResetRequestVerifications(): returned nil, but expected an error from not matching the expected verifier.", test.name)
		}
	}
}
