package vasc

import "context"

type contextKeyType struct{}

// NewContext returns a copy of ctx which carries client.
func NewContext(ctx context.Context, client *Client) context.Context {
	return context.WithValue(ctx, contextKeyType{}, client)
}

// FromContext returns the Client ctx carries.
//
// FromContext panics in case ctx carries no Client.
func FromContext(ctx context.Context) *Client {
	return ctx.Value(contextKeyType{}).(*Client)
}
