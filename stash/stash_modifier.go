package stash

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian/parse"
)

func init() {
	parse.Register("stash.Modifier", modifierFromJSON)
}

// Modifier adds a header to the request containing the current state of the URL.
// The header will be named with the value stored in headerName.
// There will be no validation done on this header name.
type Modifier struct {
	headerName string
}

type modifierJSON struct {
	HeaderName string               `json:"headerName"`
	Scope      []parse.ModifierType `json:"scope"`
}

// NewModifier returns a RequestModifier that will add a header to the request containing the current state of the URL.
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
