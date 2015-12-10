package body

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/martian/proxyutil"
)

func TestBodyFromFileModifierOnRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "test.json", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	res := proxyutil.NewResponse(200, nil, req)

	mod, err := NewFileModifier("test.json")
	if err != nil {
		t.Fatalf("NewFileModifier: got %v, want no error", err)
	}

	if err := mod.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	if got, want := res.Header.Get("Content-Type"), "application/json"; got != want {
		t.Errorf("res.Header.Get(%q): got %v, want %v", "Content-Type", got, want)
	}

	gotBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(): got %v, want no error", err)
	}

	wantBytes, err := ioutil.ReadFile("test.json")
	if err != nil {
		t.Fatalf("ioutil.ReadFile(): got %v, want no error", err)
	}

	if !bytes.Equal(gotBytes, wantBytes) {
		t.Errorf("res.Body: got %v, want %v", gotBytes, wantBytes)
	}

	if got, want := res.ContentLength, int64(len(wantBytes)); got != want {
		t.Errorf("res.Header.Get(%q): got %v, want %v", "Content-Length", got, want)
	}
}
