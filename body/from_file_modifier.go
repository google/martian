package body

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/google/martian/parse"
)

// FileModifier substitutes the body on an HTTP response with bytes read from
// a file local to the proxy.
type FileModifier struct {
	contentType string
	body        []byte
}

type fileModifierJSON struct {
	Path  string               `json:"path"`
	Scope []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("body.FileModifier", fileModifierFromJSON)
}

// NewFileModifier reads a file and constructs a modifier that will replace the
// body of an HTTP message with the contents of the file.
func NewFileModifier(path string) (*FileModifier, error) {
	ext := filepath.Ext(path)
	ct := mime.TypeByExtension(ext)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &FileModifier{
		contentType: ct,
		body:        b,
	}, nil
}

func fileModifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &fileModifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod, err := NewFileModifier(msg.Path)
	if err != nil {
		return nil, err
	}
	return parse.NewResult(mod, msg.Scope)
}

// ModifyResponse replaces the the body of an HTTP response with the bytes
// read from the file at the provided path.
func (m *FileModifier) ModifyResponse(res *http.Response) error {
	res.Body.Close()
	res.Header.Del("Content-Encoding")
	res.Header.Set("Content-Type", m.contentType)
	res.ContentLength = int64(len(m.body))
	res.Body = ioutil.NopCloser(bytes.NewReader(m.body))

	return nil
}
