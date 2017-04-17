package secure

import (
	"net/http"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/martiantest"
	"github.com/google/martian/session"
)

func TestSecureFilter(t *testing.T) {
	req, err := http.NewRequest("GET", "http://martian.local", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, err := session.FromContext(nil)
	if err != nil {
		t.Fatalf("session.FromContext(): got %v, want no error", err)
	}

	martian.SetContext(req, ctx)

	session := ctx.GetSession()

	f := NewSecureFilter()
	tm := martiantest.NewModifier()
	f.SetRequestModifier(tm)

	if err := f.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got := tm.RequestModified(); got == true {
		t.Errorf("tm.RequestModified(): got %v, want false", got)
	}

	session.MarkSecure()

	tm = martiantest.NewModifier()
	f.SetRequestModifier(tm)

	if err := f.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got := tm.RequestModified(); got == false {
		t.Errorf("tm.RequestModified(): got %v, want true", got)
	}
}

func TestUnsecureFilter(t *testing.T) {
	req, err := http.NewRequest("GET", "http://martian.local", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, err := session.FromContext(nil)
	if err != nil {
		t.Fatalf("session.FromContext(): got %v, want no error", err)
	}

	martian.SetContext(req, ctx)

	session := ctx.GetSession()

	f := NewUnsecureFilter()
	tm := martiantest.NewModifier()
	f.SetRequestModifier(tm)

	if err := f.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got := tm.RequestModified(); got == false {
		t.Errorf("tm.RequestModified(): got %v, want true", got)
	}

	session.MarkSecure()

	tm = martiantest.NewModifier()
	f.SetRequestModifier(tm)

	if err := f.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	if got := tm.RequestModified(); got == true {
		t.Errorf("tm.RequestModified(): got %v, want false", got)
	}
}
