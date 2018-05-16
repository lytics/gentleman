package proxy

import (
	"net/http"
	"testing"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/utils"
)

func TestProxy(t *testing.T) {
	ctx := context.New()
	ctx.Request.URL.Scheme = "http"

	fn := newHandler()
	servers := map[string]string{"http": "http://localhost:3128"}

	Set(servers).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	transport := ctx.Client.Transport.(*http.Transport)
	url, err := transport.Proxy(ctx.Request)

	utils.Equal(t, err, nil)
	utils.Equal(t, url.Host, "localhost:3128")
	utils.Equal(t, url.Scheme, "http")
}

func TestProxyParseError(t *testing.T) {
	ctx := context.New()
	ctx.Request.URL.Scheme = "http"

	fn := newHandler()
	servers := map[string]string{"http": "://"}

	Set(servers).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	transport := ctx.Client.Transport.(*http.Transport)
	_, err := transport.Proxy(ctx.Request)

	utils.Equal(t, err.Error(), "parse ://: missing protocol scheme")
}

type handler struct {
	fn     context.Handler
	called bool
}

func newHandler() *handler {
	h := &handler{}
	h.fn = context.NewHandler(func(c *context.Context) {
		h.called = true
	})
	return h
}
