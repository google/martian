// Copyright 2023 Sauce Labs, Inc. All rights reserved.

package dialvia

import (
	"context"
	"net"
)

// ContextDialerFunc is a function that implements Dialer and ContextDialer.
type ContextDialerFunc func(context context.Context, network, addr string) (net.Conn, error)

// Dial is needed to satisfy the proxy.Dialer interface.
// It is never called as proxy.ContextDialer is used instead if available.
func (f ContextDialerFunc) Dial(network, addr string) (net.Conn, error) {
	return f(context.Background(), network, addr)
}

func (f ContextDialerFunc) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return f(ctx, network, addr)
}
