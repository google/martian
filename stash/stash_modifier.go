package stash

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian/parse"
)

func init() {
	parse.Register("stash.Modifier", modifierFromJSON)
}

// Modifier alters the request URL and Host header to
type Modifier struct {
	headerName string
}

type modifierJSON struct {
	HeaderName string               `json:"headerName"`
	Scope      []parse.ModifierType `json:"scope"`
}

// NewModifier returns a RequestModifier that can be configured to
func NewModifier(headerName string) *Modifier {
	return &Modifier{headerName: headerName}
}

// ModifyRequest alters the request
// See docs for Modifier for details.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	req.Header.Set(m.headerName, req.URL.String())
	return nil
}

func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewModifier(msg.HeaderName)
	return parse.NewResult(mod, msg.Scope)
}
