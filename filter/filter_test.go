package filter

import (
	"net/http"
	"testing"

	"github.com/google/martian/header"
	"github.com/google/martian/martiantest"
)

func TestRequestWhenTrueCondition(t *testing.T) {
	hm := header.NewMatcher("Martian-Testing", "true")

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
		tm := martiantest.NewModifier()

		f := New()
		f.SetRequestCondition(hm)
		f.RequestWhenTrue(tm)

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		req.Header[tc.name] = tc.values

		if err := f.ModifyRequest(req); err != nil {
			t.Fatalf("%d. ModifyRequest(): got %v, want no error", i, err)
		}

		if tm.RequestModified() != tc.want {
			t.Errorf("%d. tm.RequestModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}

func TestRequestWhenFalse(t *testing.T) {
	hm := header.NewMatcher("Martian-Testing", "true")
	tt := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "Martian-Production",
			values: []string{"true"},
			want:   true,
		},
		{
			name:   "Martian-Testing",
			values: []string{"see-next-value", "true"},
			want:   false,
		},
	}

	for i, tc := range tt {
		tm := martiantest.NewModifier()

		f := New()
		f.SetRequestCondition(hm)
		f.RequestWhenFalse(tm)

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		req.Header[tc.name] = tc.values

		if err := f.ModifyRequest(req); err != nil {
			t.Fatalf("%d. ModifyRequest(): got %v, want no error", i, err)
		}

		if tm.RequestModified() != tc.want {
			t.Errorf("%d. tm.RequestModified(): got %t, want %t", i, tm.RequestModified(), tc.want)
		}
	}
}
