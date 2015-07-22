package martianhttp

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/martian/mitm"
)

func TestAuthorityHandler(t *testing.T) {
	ca, _, err := mitm.NewAuthority("martian.proxy", "Martian Authority", time.Hour)
	if err != nil {
		t.Fatalf("mitm.NewAuthority(): got %v, want no error", err)
	}

	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/martian/authority.cer", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	h := NewAuthorityHandler(ca)
	h.ServeHTTP(rw, req)

	if got, want := rw.Code, 200; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}
	if got, want := rw.Header().Get("Content-Type"), "application/x-x509-ca-cert"; got != want {
		t.Errorf("rw.Header().Get(%q): got %q, want %q", "Content-Type", got, want)
	}

	blk, _ := pem.Decode(rw.Body.Bytes())
	if got, want := blk.Type, "CERTIFICATE"; got != want {
		t.Errorf("rw.Body: got PEM type %q, want %q", got, want)
	}

	cert, err := x509.ParseCertificate(blk.Bytes)
	if err != nil {
		t.Fatalf("x509.ParseCertificate(res.Body): got %v, want no error", err)
	}
	if got, want := cert.Subject.CommonName, "martian.proxy"; got != want {
		t.Errorf("cert.Subject.CommonName: got %q, want %q", got, want)
	}
}
