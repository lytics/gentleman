package query

import (
	"net/url"

	c "github.com/lytics/gentleman/context"
	p "github.com/lytics/gentleman/plugin"
)

// Set sets the query param key and value.
// It replaces any existing values.
func Set(key, value string) p.Plugin {
	return p.NewRequestPlugin(func(ctx *c.Context, h c.Handler) {
		query := ctx.Request.URL.Query()
		query.Set(key, value)
		ctx.Request.URL.RawQuery = query.Encode()
		h.Next(ctx)
	})
}

// Add adds the query param value to key.
// It appends to any existing values associated with key.
func Add(key, value string) p.Plugin {
	return p.NewRequestPlugin(func(ctx *c.Context, h c.Handler) {
		query := ctx.Request.URL.Query()
		query.Add(key, value)
		ctx.Request.URL.RawQuery = query.Encode()
		h.Next(ctx)
	})
}

// Del deletes the query param values associated with key.
func Del(key string) p.Plugin {
	return p.NewRequestPlugin(func(ctx *c.Context, h c.Handler) {
		query := ctx.Request.URL.Query()
		query.Del(key)
		ctx.Request.URL.RawQuery = query.Encode()
		h.Next(ctx)
	})
}

// DelAll deletes all the query params.
func DelAll() p.Plugin {
	return p.NewRequestPlugin(func(ctx *c.Context, h c.Handler) {
		ctx.Request.URL.RawQuery = ""
		h.Next(ctx)
	})
}

// SetMap sets a map of query params by key-value pair.
func SetMap(params url.Values) p.Plugin {
	return p.NewRequestPlugin(func(ctx *c.Context, h c.Handler) {
		query := ctx.Request.URL.Query()
		for key, values := range params {
			for _, v := range values {
				query.Set(key, v)
			}
		}
		ctx.Request.URL.RawQuery = query.Encode()
		h.Next(ctx)
	})
}
