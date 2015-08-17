package martianlua

import (
	"io/ioutil"
	"net/http"

	"github.com/google/martian"
)

type handler struct {
	m *Modifier
}

func NewHandler(m *Modifier) http.Handler {
	return &handler{
		m: m,
	}
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		martian.Errorf("martianlua: invalid request method: %s", req.Method)
		rw.Header().Set("Allow", "POST")
		http.WriteHeader(405)
		return
	}

	script, err := ioutil.ReadAll(req.Body)
	if err != nil {
		martian.Errorf("martianlua: failed to read request body: %v", err)
		http.Error(rw, err.Error(), 500)
		return
	}

	m.SetScript(string(script))
}
