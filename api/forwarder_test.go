package api

import (
	"net/http"
	"testing"

	"github.com/google/martian"
)

func TestApiForwarder(t *testing.T) {
	forwarder := NewForwarder("apihost.com")

	req, err := http.NewRequest("GET", "https://localhost:8080/configure", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}

	ctx, remove, err := martian.TestContext(req, nil, nil)
	if err != nil {
		t.Fatalf("TestContext(): got %v, want no error", err)
	}
	defer remove()

	if err := forwarder.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got, want := req.URL.Scheme, "http"; got != want {
		t.Errorf("req.URL.Scheme: got %s, want %s", got, want)
	}
	if got, want := req.URL.Host, "apihost.com"; got != want {
		t.Errorf("req.URL.Host: got %s, want %s", got, want)
	}

	if !ctx.SkippingLogging() {
		t.Errorf("SkippingLogging: got false, want true")
	}
}
