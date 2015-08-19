package verify

import "fmt"

type MessageType string

const (
	// Request signifies that the error occurred on a request.
	Request MessageType = "request"
	// Response signifies that the error occurred on a response.
	Response MessageType = "response"
)

type Error struct {
	Kind     string      `json:"kind"`
	Scope    MessageType `json:"scope"`
	URL      string      `json:"url"`
	Actual   string      `json:"actual"`
	Expected string      `json:"expected"`
	Message  string      `json:"message"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("%s: %s(%s): got %q, want %q", e.Kind, e.Scope, e.URL, actual, e.Expected)
	}

	return e.Message
}
