package transport

import (
	"net/http"
	"testing"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/utils"
)

func TestSetTransport(t *testing.T) {
	ctx := context.New()
	fn := newHandler()
	transport := &http.Transport{}
	Set(transport).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	newTransport := ctx.Client.Transport.(*http.Transport)
	utils.Equal(t, newTransport, transport)
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
