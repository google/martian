package martianlua

import (
	"net/http"
	"testing"
)

func TestModifier(t *testing.T) {
	m := NewModifier()
	m.SetScript(`
		local httpenforcer = {}

		function httpenforcer.modifyrequest(request)
			request.url = "https://example.com"
		end

		return httpenforcer
	`)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}
	if err := m.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got, want := req.URL.Scheme, "https"; got != want {
		t.Errorf("req.URL.Scheme: got %q, want %q", got, want)
	}
}
