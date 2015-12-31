package verify

import (
	"fmt"
	"net/http"
	"net/url"
)

// Scope describes whether the verification error corresponds to a
// request or a response.
type Scope uint8

const (
	// Unknown signifies the error occurred on an unknown message type (default).
	Unknown Scope = iota
	// Request signifies the error occurred on a request.
	Request Scope = iota
	// Response signifies the error occurred on a response.
	Response Scope = iota
)

// ErrorFormat is the format that will be used for verify.Error message
// strings. The order of the four string format verbs is: Kind, Scope, URL,
// MessageFormat.
//
// MessageFormat is expected to be a format string containing two string format
// verbs in the following order: Actual, Expected.
var ErrorFormat = "%s: %s(%s): %s"

// Error is a concrete verification error.
type Error struct {
	Kind     string
	Scope    Scope
	URL      string
	Actual   string
	Expected string
	Message  string
}

type ErrorBuilder struct {
	kind      string
	scope     Scope
	url       string
	actual    string
	expected  string
	formatted bool
	message   string
	cond      func() bool
	reset     func()
}

func NewError(kind string) *ErrorBuilder {
	return &ErrorBuilder{
		kind:  kind,
		cond:  func() bool { return true },
		reset: func() {},
	}
}

func (eb *ErrorBuilder) Request(req *http.Request) *ErrorBuilder {
	eb.scope = Request
	eb.url = req.URL.String()

	return eb
}

func (eb *ErrorBuilder) Response(res *http.Response) *ErrorBuilder {
	eb.scope = Response
	eb.url = res.Request.URL.String()

	return eb
}

func (eb *ErrorBuilder) For(scope Scope, u *url.URL) *ErrorBuilder {
	eb.scope = scope
	eb.url = u.String()

	return eb
}

func (eb *ErrorBuilder) Conditionally(cond func() bool) *ErrorBuilder {
	eb.cond = cond

	return eb
}

func (eb *ErrorBuilder) Resets(reset func()) *ErrorBuilder {
	eb.reset = reset

	return eb
}

func (eb *ErrorBuilder) Message(message string) *ErrorBuilder {
	eb.message = message

	return eb
}

func (eb *ErrorBuilder) Format(format string) *ErrorBuilder {
	eb.formatted = true
	eb.message = format

	return eb
}

func (eb *ErrorBuilder) Actual(actual string) *ErrorBuilder {
	eb.actual = actual

	return eb
}

func (eb *ErrorBuilder) Expected(expected string) *ErrorBuilder {
	eb.expected = expected

	return eb
}

func (eb *ErrorBuilder) Error() (Error, bool) {
	if !eb.cond() {
		return Error{}, false
	}

	return Error{
		Kind:     eb.kind,
		Scope:    eb.scope,
		URL:      eb.url,
		Actual:   eb.actual,
		Expected: eb.expected,
		Message:  eb.formattedMessage(),
	}, true
}

func (eb *ErrorBuilder) formattedMessage() string {
	switch {
	case !eb.formatted:
		return eb.message
	case eb.actual == "" && eb.expected == "":
		return "failed verification"
	case eb.actual == "":
		return fmt.Sprintf(eb.message, eb.expected)
	}

	return fmt.Sprintf(eb.message, eb.actual, eb.expected)
}

func (eb *ErrorBuilder) Reset() {
	eb.reset()
}

// String returns the string represenation of the scope.
func (s Scope) String() string {
	switch s {
	case Request:
		return "request"
	case Response:
		return "response"
	default:
		return "unknown"
	}
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Scope) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case "request":
		*s = Request
	case "response":
		*s = Response
	default:
		*s = Unknown
	}

	return nil
}
