package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// MessageType describes whether the verification error corresponds to a
// request or a response.
type MessageType uint8

// ErrorValue is a concrete verification error.
type ErrorValue struct {
	Kind          string
	Scope         MessageType
	URL           string
	Actual        string
	Expected      string
	MessageFormat string
}

const (
	// Unknown signifies the error occurred on an unknown message type (default).
	Unknown MessageType = iota
	// Request signifies the error occurred on a request.
	Request MessageType = iota
	// Response signifies the error occurred on a response.
	Response MessageType = iota
)

// RequestError returns a verification error for the given request.
func RequestError(kind string, req *http.Request) *ErrorValue {
	return &ErrorValue{
		Kind:          kind,
		Scope:         Request,
		URL:           req.URL.String(),
		MessageFormat: "got %s, want %s",
	}
}

// ResponseError returns a verification error for the given response.
func ResponseError(kind string, res *http.Response) *ErrorValue {
	return &ErrorValue{
		Kind:          kind,
		Scope:         Response,
		URL:           res.Request.URL.String(),
		MessageFormat: "got %s, want %s",
	}
}

// MarshalJSON builds a JSON representation of a verification error.
func (ev *ErrorValue) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	m := map[string]string{
		"kind":    ev.Kind,
		"scope":   ev.Scope.String(),
		"url":     ev.URL,
		"message": ev.Error(),
	}

	if ev.Actual != "" {
		m["actual"] = ev.Actual
	}

	if ev.Expected != "" {
		m["expected"] = ev.Expected
	}

	json.NewEncoder(buf).Encode(m)

	return buf.Bytes(), nil
}

// Get returns itself so it can be used to satisfy the verify.Error interface.
func (ev *ErrorValue) Get() *ErrorValue {
	return ev
}

// Error returns the error string.
func (ev *ErrorValue) Error() string {
	msg := fmt.Sprintf("%s: %s(%s): %s", ev.Kind, ev.Scope, ev.URL, ev.MessageFormat)

	switch {
	case ev.Actual == "" && ev.Expected == "":
		return fmt.Sprintf(msg, "failure", "success")
	case ev.Actual == "":
		return fmt.Sprintf(msg, ev.Expected)
	}

	return fmt.Sprintf(msg, ev.Actual, ev.Expected)
}

// String returns the string represenation of the message type.
func (mt MessageType) String() string {
	switch mt {
	case Request:
		return "request"
	case Response:
		return "response"
	default:
		return "unknown"
	}
}
