package skip

import (
	"net/http"
	"testing"

	"github.com/google/martian/martian"
	"github.com/google/martian/parse/parse"
)

func TestRoundTrip(t *testing.T) {
	m := NewRoundTrip()
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req)
	if err != nil {
		t.Fatalf("martian.TestContext(): got %v, want no error", err)
	}
	defer remove()

	if ctx.SkippingRoundTrip() {
		t.Fatal("ctx.SkippingRoundTrip(): got true, want false")
	}

	if err := m.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if !ctx.SkippingRoundTrip() {
		t.Fatal("ctx.SkippingRoundTrip(): got false, want true")
	}
}

func TestFromJSON(t *testing.T) {
	msg := []byte(`{
			  "skip.RoundTrip": {}
	        }`)

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got %v, want no error", err)
	}

	reqmod := r.RequestModifier()
	if _, ok := reqmod.(*RoundTrip); !ok {
		t.Fatal("reqmod.(*RoundTrip): got !ok, want ok")
	}
}
