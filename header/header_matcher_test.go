package header

import (
	"net/http"
	"testing"

	"github.com/google/martian/proxyutil"
)

func TestMatchResponse(t *testing.T) {
	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   false,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   true,
		},
	}

	for i, tc := range tt {
		matcher := NewMatcher("Martian-Testing", "true")
		res := proxyutil.NewResponse(200, nil, nil)
		res.Header[tc.name] = tc.values

		if got := matcher.MatchResponse(res); got != tc.want {
			t.Fatalf("%d. MatchResponse(): got %t, want %t", i, got, tc.want)
		}
	}
}

func TestMatchRequest(t *testing.T) {
	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   false,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   true,
		},
	}

	for i, tc := range tt {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		matcher := NewMatcher("Martian-Testing", "true")
		req.Header[tc.name] = tc.values

		if got := matcher.MatchRequest(req); got != tc.want {
			t.Fatalf("%d. MatchRequest(): got %t, want %t", i, got, tc.want)
		}
	}
}
