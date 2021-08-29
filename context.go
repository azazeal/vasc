package vasc

import "context"

type contextKeyType struct{}

// NewContext returns a copy of the given Context which carries client.
func NewContext(ctx context.Context, client *Client) context.Context {
	return context.WithValue(ctx, contextKeyType{}, client)
}

// FromContext returns the Client the given Context carries.
//
// FromContext panics in case the given Context carries no Check.
func FromContext(ctx context.Context) *Client {
	return ctx.Value(contextKeyType{}).(*Client)
}
