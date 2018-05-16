package compression

import (
	"net/http"
	"testing"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/utils"
)

func TestDisableCompression(t *testing.T) {
	ctx := context.New()
	fn := newHandler()
	Disable().Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	transport := ctx.Client.Transport.(*http.Transport)
	utils.Equal(t, transport.DisableCompression, true)
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
