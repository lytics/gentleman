package transport

import (
	c "github.com/lytics/gentleman/context"
	p "github.com/lytics/gentleman/plugin"
	"net/http"
)

// Set sets a new HTTP transport for the outgoing request
func Set(transport http.RoundTripper) p.Plugin {
	return p.NewRequestPlugin(func(ctx *c.Context, h c.Handler) {
		// Override the http.Client transport
		ctx.Client.Transport = transport
		h.Next(ctx)
	})
}
