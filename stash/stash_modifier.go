package stash

import (
	"encoding/json"
	"fmt"
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

// NewModifier returns a RequestModifier that write the current URL into a header.
func NewModifier(headerName string) *Modifier {
	return &Modifier{headerName: headerName}
}

// ModifyRequest writes the current URL into a header.
// See docs for Modifier for details.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	req.Header.Set(m.headerName, req.URL.String())
	return nil
}

// ModifyResponse writes the same header written in the request into the response.
func (m *Modifier) ModifyResponse(res *http.Response) error {
	res.Header.Set(m.headerName, res.Request.Header.Get(m.headerName))
	return nil
}

// If you would like the saved state of the URL to be written in the response you must specify this modifier's scope as both request and response.
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewModifier(msg.HeaderName)
	result, err := parse.NewResult(mod, msg.Scope)

	if result.ResponseModifier() != nil && result.RequestModifier() == nil {
		return nil, fmt.Errorf("To write header on a response, specify scope as both request and response.")
	}

	return result, err
}
