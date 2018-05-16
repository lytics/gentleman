package timeout

import (
	"net/http"
	"testing"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/utils"
)

func TestTimeout(t *testing.T) {
	ctx := context.New()
	fn := newHandler()
	Request(1000).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	utils.Equal(t, ctx.Error, nil)
	utils.Equal(t, int(ctx.Client.Timeout), 1000)
}

func TestTimeoutTLS(t *testing.T) {
	ctx := context.New()
	fn := newHandler()
	TLS(1000).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	utils.Equal(t, ctx.Error, nil)

	transport := ctx.Client.Transport.(*http.Transport)
	utils.Equal(t, int(transport.TLSHandshakeTimeout), 1000)
}

func TestTimeoutAll(t *testing.T) {
	ctx := context.New()
	fn := newHandler()
	All(Timeouts{Request: 1000, Dial: 1000, TLS: 1000}).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	utils.Equal(t, ctx.Error, nil)
	transport := ctx.Client.Transport.(*http.Transport)
	utils.Equal(t, int(ctx.Client.Timeout), 1000)
	utils.Equal(t, int(transport.TLSHandshakeTimeout), 1000)
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
