// Package skip provides a request modifier to skip the HTTP round-trip.
package skip

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian"
	"github.com/google/martian/parse"
)

// RoundTrip is a modifier that skips the request round-trip.
type RoundTrip struct{}

type roundTripJSON struct {
	Scope []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("skip.RoundTrip", roundTripFromJSON)
}

// NewRoundTrip returns a new modifier that skips round-trip.
func NewRoundTrip() *RoundTrip {
	return &RoundTrip{}
}

// ModifyRequest skips the request round-trip.
func (r *RoundTrip) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	ctx.SkipRoundTrip()

	return nil
}

// roundTripFromJSON builds a skip.RoundTrip from JSON.

// Example JSON:
// {
//   "skip.RoundTrip": { }
// }
func roundTripFromJSON(b []byte) (*parse.Result, error) {
	msg := &roundTripJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewRoundTrip(), msg.Scope)
}
