package body

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/utils"
)

func TestBodyJSONEncodeMap(t *testing.T) {
	ctx := context.New()
	fn := newHandler()

	json := map[string]string{"foo": "bar"}
	JSON(json).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "GET")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "application/json")
	utils.Equal(t, int(ctx.Request.ContentLength), 14)
	utils.Equal(t, string(buf[0:len(buf)-1]), `{"foo":"bar"}`)
}

func TestBodyJSONEncodeString(t *testing.T) {
	ctx := context.New()
	fn := newHandler()

	json := `{"foo":"bar"}`
	JSON(json).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "GET")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "application/json")
	utils.Equal(t, int(ctx.Request.ContentLength), 13)
	utils.Equal(t, string(buf), `{"foo":"bar"}`)
}

func TestBodyJSONEncodeBytes(t *testing.T) {
	ctx := context.New()
	fn := newHandler()

	json := []byte(`{"foo":"bar"}`)
	JSON(json).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)
	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "GET")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "application/json")
	utils.Equal(t, int(ctx.Request.ContentLength), 13)
	utils.Equal(t, string(buf), `{"foo":"bar"}`)
}

func TestBodyXMLEncodeStruct(t *testing.T) {
	ctx := context.New()
	fn := newHandler()

	type xmlTest struct {
		Name string `xml:"name>first"`
	}
	xml := xmlTest{Name: "foo"}
	XML(xml).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "GET")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "application/xml")
	utils.Equal(t, int(ctx.Request.ContentLength), 50)
	utils.Equal(t, string(buf), `<xmlTest><name><first>foo</first></name></xmlTest>`)
}

func TestBodyXMLEncodeString(t *testing.T) {
	ctx := context.New()
	fn := newHandler()

	xml := "<test>foo</test>"
	XML(xml).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "GET")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "application/xml")
	utils.Equal(t, int(ctx.Request.ContentLength), 16)
	utils.Equal(t, string(buf), `<test>foo</test>`)
}

func TestBodyXMLEncodeBytes(t *testing.T) {
	ctx := context.New()
	ctx.Request.Method = ""
	fn := newHandler()

	xml := []byte("<test>foo</test>")
	XML(xml).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "POST")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "application/xml")
	utils.Equal(t, int(ctx.Request.ContentLength), 16)
	utils.Equal(t, string(buf), `<test>foo</test>`)
}

func TestBodyReader(t *testing.T) {
	ctx := context.New()
	ctx.Request.Method = "POST"
	fn := newHandler()

	reader := bytes.NewReader([]byte("foo bar"))
	Reader(reader).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "POST")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "")
	utils.Equal(t, int(ctx.Request.ContentLength), 7)
	utils.Equal(t, string(buf), "foo bar")
}

func TestBodyReaderContextDataSharing(t *testing.T) {
	ctx := context.New()
	ctx.Request.Method = "POST"
	fn := newHandler()

	// Set sample context data
	ctx.Set("foo", "bar")
	ctx.Set("bar", "baz")

	reader := bytes.NewReader([]byte("foo bar"))
	Reader(reader).Exec("request", ctx, fn.fn)
	utils.Equal(t, fn.called, true)

	buf, err := ioutil.ReadAll(ctx.Request.Body)
	utils.Equal(t, err, nil)
	utils.Equal(t, ctx.Request.Method, "POST")
	utils.Equal(t, ctx.Request.Header.Get("Content-Type"), "")
	utils.Equal(t, int(ctx.Request.ContentLength), 7)
	utils.Equal(t, string(buf), "foo bar")

	// Test context data
	utils.Equal(t, ctx.GetString("foo"), "bar")
	utils.Equal(t, ctx.GetString("bar"), "baz")
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
