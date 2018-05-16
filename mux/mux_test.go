package mux

import (
	"testing"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/utils"
)

func TestMuxSimple(t *testing.T) {
	mx := New()
	mx.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.Header.Set("foo", "bar")
		h.Next(ctx)
	})
	ctx := context.New()
	mx.Run("request", ctx)
	utils.Equal(t, ctx.Request.Header.Get("foo"), "bar")
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
