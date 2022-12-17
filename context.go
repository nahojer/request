package request

import (
	"context"
	"net/http"
)

// contextKey represents the type of value for the context key.
type contextKey string

// clientContextKey is how *http.Client is stored/retrieved in context.
const clientContextKey contextKey = "clientContextKey"

// AttachClientToContext attaches c to ctx. It allows for overriding the
// default HTTP client used by [Request.Do].
func AttachClientToContext(ctx context.Context, c *http.Client) context.Context {
	return context.WithValue(ctx, clientContextKey, c)
}

func clientFromContext(ctx context.Context) *http.Client {
	c, ok := ctx.Value(clientContextKey).(*http.Client)
	if !ok || c == nil {
		return &http.Client{
			Timeout: DefaultClientTimeout,
		}
	}
	return c
}
