package tls

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/utils"
)

func TestAuthBasic(t *testing.T) {
	ctx := context.New()
	fn := newHandler()
	config := &tls.Config{InsecureSkipVerify: true}
	Config(config).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	transport := ctx.Client.Transport.(*http.Transport)
	utils.Equal(t, transport.TLSClientConfig, config)
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
