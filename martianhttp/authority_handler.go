package martianhttp

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
)

type authorityHandler struct {
	cert []byte
}

// NewAuthorityHandler returns an http.Handler that will present the client
// with the CA certificate to use in browser.
func NewAuthorityHandler(ca *x509.Certificate) http.Handler {
	return &authorityHandler{
		cert: pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: ca.Raw,
		}),
	}
}

// ServeHTTP writes the CA certificate in PEM format to the client.
func (h *authorityHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/x-x509-ca-cert")
	rw.Write(h.cert)
}
