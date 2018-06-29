package gentleman

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/lytics/gentleman/context"

	// "github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/middleware"
	"github.com/lytics/gentleman/mux"
	"github.com/lytics/gentleman/plugin"
	"github.com/lytics/gentleman/plugins/body"
	"github.com/lytics/gentleman/plugins/bodytype"
	"github.com/lytics/gentleman/plugins/cookies"
	"github.com/lytics/gentleman/plugins/headers"
	"github.com/lytics/gentleman/plugins/multipart"
	"github.com/lytics/gentleman/plugins/query"
	u "github.com/lytics/gentleman/plugins/url"
)

const (
	// UserAgent represents the static user agent name and version.
	UserAgent = "gentleman/" + Version
)

var (
	// DialTimeout represents the maximum amount of time the network dialer can take.
	DialTimeout = 30 * time.Second

	// DialKeepAlive represents the maximum amount of time too keep alive the socket.
	DialKeepAlive = 30 * time.Second

	// TLSHandshakeTimeout represents the maximum amount of time that
	// TLS handshake can take defined in the default http.Transport.
	TLSHandshakeTimeout = 10 * time.Second

	// RequestTimeout represents the maximum about of time that
	// a request can take, including dial / request / redirect processes.
	RequestTimeout = 60 * time.Second

	// DefaultDialer defines the default network dialer.
	DefaultDialer = &net.Dialer{
		Timeout:   DialTimeout,
		KeepAlive: DialKeepAlive,
	}

	// DefaultTransport stores the default HTTP transport to be used.
	DefaultTransport = NewDefaultTransport(DefaultDialer)
)

// Request HTTP entity for gentleman.
// Provides middleware capabilities, built-in context
// and convenient methods to easily setup request params.
type Request struct {
	// Stores if the request was already dispatched
	dispatched bool

	// Optional reference to the gentleman.Client instance
	Client *Client

	// Request scope Context instance
	Context *context.Context

	// Request scope Middleware instance
	Middleware middleware.Middleware
}

// NewRequest creates a new Request entity.
func NewRequest() *Request {
	ctx := context.New()
	ctx.Client.Transport = DefaultTransport
	ctx.Request.Header.Set("User-Agent", UserAgent)
	return &Request{
		Context:    ctx,
		Middleware: middleware.New(),
	}
}

// SetClient Attach a client to the current Request
// This is mostly done internally.
func (r *Request) SetClient(cli *Client) *Request {
	r.Client = cli
	r.Context.UseParent(cli.Context)
	r.Middleware.UseParent(cli.Middleware)
	return r
}

// Mux is a middleware multiplexer for easy plugin composition.
func (r *Request) Mux() *mux.Mux {
	mx := mux.New()
	r.Use(mx)
	return mx
}

// Method defines the HTTP verb to be used.
func (r *Request) Method(method string) *Request {
	r.Middleware.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.Method = method
		h.Next(ctx)
	})
	return r
}

// URL parses and defines the URL to be used in the outgoing request.
func (r *Request) URL(uri string) *Request {
	r.Use(u.URL(uri))
	return r
}

// BaseURL parses the given URL and uses the URL schema and host in the outgoing request.
func (r *Request) BaseURL(uri string) *Request {
	r.Use(u.BaseURL(uri))
	return r
}

// Path defines the request URL path to be used in the outgoing request.
func (r *Request) Path(path string) *Request {
	r.Use(u.Path(path))
	return r
}

// AddPath defines the request URL path to be used in the outgoing request.
func (r *Request) AddPath(path string) *Request {
	r.Use(u.AddPath(path))
	return r
}

// Param replaces a path param based on the given param name and value.
func (r *Request) Param(name, value string) *Request {
	r.Use(u.Param(name, value))
	return r
}

// Params replaces path params based on the given params key-value map.
func (r *Request) Params(params map[string]string) *Request {
	r.Use(u.Params(params))
	return r
}

// SetQuery sets a new URL query param field.
// If another query param exists with the same key, it will be overwritten.
func (r *Request) SetQuery(name, value string) *Request {
	r.Use(query.Set(name, value))
	return r
}

// AddQuery adds a new URL query param field
// without overwriting any existent query field.
func (r *Request) AddQuery(name, value string) *Request {
	r.Use(query.Add(name, value))
	return r
}

// SetQueryParams sets URL query params based on the given map.
func (r *Request) SetQueryParams(params url.Values) *Request {
	r.Use(query.SetMap(params))
	return r
}

// SetHeader sets a new header field by name and value.
// If another header exists with the same key, it will be overwritten.
func (r *Request) SetHeader(name, value string) *Request {
	r.Use(headers.Set(name, value))
	return r
}

// AddHeader adds a new header field by name and value
// without overwriting any existent header.
func (r *Request) AddHeader(name, value string) *Request {
	r.Use(headers.Add(name, value))
	return r
}

// SetHeaders adds new header fields based on the given map.
func (r *Request) SetHeaders(fields map[string]string) *Request {
	r.Use(headers.SetMap(fields))
	return r
}

// AddCookie sets a new cookie field bsaed on the given http.Cookie struct
// without overwriting any existent cookie.
func (r *Request) AddCookie(cookie *http.Cookie) *Request {
	r.Use(cookies.Add(cookie))
	return r
}

// AddCookies sets a new cookie field based on a list of http.Cookie
// without overwriting any existent cookie.
func (r *Request) AddCookies(data []*http.Cookie) *Request {
	r.Use(cookies.AddMultiple(data))
	return r
}

// CookieJar creates a cookie jar to store HTTP cookies when they are sent down.
// func (r *Request) CookieJar() *Request {
// 	r.Use(cookies.Jar())
// 	return r
// }

// Type defines the Content-Type header field based on the given type name alias or value.
// You can use the following content type aliases: json, xml, form, html, text and urlencoded.
func (r *Request) Type(name string) *Request {
	r.Use(bodytype.Set(name))
	return r
}

// Body defines the request body based on a io.Reader stream.
func (r *Request) Body(reader io.Reader) *Request {
	r.Use(body.Reader(reader))
	return r
}

// BodyString defines the request body based on the given string.
// If using this method, you should define the proper Content-Type header
// representing the real content MIME type.
func (r *Request) BodyString(data string) *Request {
	r.Use(body.String(data))
	return r
}

// JSON serializes and defines as request body based on the given input.
// The proper Content-Type header will be transparently added for you.
func (r *Request) JSON(data interface{}) *Request {
	r.Use(body.JSON(data))
	return r
}

// XML serializes and defines the request body based on the given input.
// The proper Content-Type header will be transparently added for you.
func (r *Request) XML(data interface{}) *Request {
	r.Use(body.XML(data))
	return r
}

// Form serializes and defines the request body as multipart/form-data
// based on the given form data.
func (r *Request) Form(data multipart.FormData) *Request {
	r.Use(multipart.Data(data))
	return r
}

// File serializes and defines the request body as multipart/form-data
// containing one file field.
func (r *Request) File(name string, reader io.Reader) *Request {
	r.Use(multipart.File(name, reader))
	return r
}

// Files serializes and defines the request body as multipart/form-data
// containing the given file fields.
func (r *Request) Files(files []multipart.FormFile) *Request {
	r.Use(multipart.Files(files))
	return r
}

// Send is an alias to Do(), which executes the current request
// and returns the response.
func (r *Request) Send() (*Response, error) {
	return r.Do()
}

// Do performs the HTTP request and returns the HTTP response.
func (r *Request) Do() (*Response, error) {
	if r.dispatched {
		return nil, errors.New("gentleman: Request was already dispatched")
	}

	r.dispatched = true
	ctx := NewDispatcher(r).Dispatch()

	return buildResponse(ctx)
}

// Use uses a new plugin in the middleware stack.
func (r *Request) Use(p plugin.Plugin) *Request {
	r.Middleware.Use(p)
	return r
}

// UseRequest uses a request middleware handler.
func (r *Request) UseRequest(fn context.HandlerFunc) *Request {
	r.Middleware.UseRequest(fn)
	return r
}

// UseResponse uses a response middleware handler.
func (r *Request) UseResponse(fn context.HandlerFunc) *Request {
	r.Middleware.UseResponse(fn)
	return r
}

// UseError uses an error middleware handler.
func (r *Request) UseError(fn context.HandlerFunc) *Request {
	r.Middleware.UseError(fn)
	return r
}

// UseHandler uses an new middleware handler for the given phase.
func (r *Request) UseHandler(phase string, fn context.HandlerFunc) *Request {
	r.Middleware.UseHandler(phase, fn)
	return r
}

// Clone creates a new side-effects free Request based on the current one.
func (r *Request) Clone() *Request {
	req := NewRequest()
	req.Client = r.Client
	req.Context = r.Context.Clone()
	req.Middleware = r.Middleware.Clone()
	return req
}

// NewDefaultTransport returns a new http.Transport with default values
// based on the given net.Dialer.
func NewDefaultTransport(dialer *net.Dialer) *http.Transport {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: TLSHandshakeTimeout,
	}
	return transport
}

// Dump returns the given request in its HTTP/1.x wire
// representation. The request is Cloned, middlewares set to run before
// request and dial are run on the clone. The resulting request is
// passed to httputils.DumpRequest for the HTTP1.x wire representation.
//
// If body is true, Dump also returns the body. If Dump returns an error,
// the state of req is undefined.
//
// The documentation for httputils.DumpRequest details the returned
// representation
func (r *Request) Dump(body bool) ([]byte, error) {
	tmp := r.Clone()
	pipeline := []task{
		func(ctx *context.Context) (*context.Context, bool) {
			return tmp.run("request", ctx)
		},
		func(ctx *context.Context) (*context.Context, bool) {
			return tmp.run("before dial", ctx)
		},
	}
	// Reference to initial context
	ctx := tmp.Context
	// Execute tasks in order, stopping in case of error or explicit stop.
	for _, task := range pipeline {
		var stop bool
		if ctx, stop = task(ctx); stop {
			break
		}
	}
	b, err := httputil.DumpRequest(ctx.Request, body)
	return b, err
}

// run a middleware phase
func (r *Request) run(phase string, ctx *context.Context) (*context.Context, bool) {
	mw := r.Middleware

	// Run the middleware by phase
	ctx = mw.Run(phase, ctx)
	if ctx.Error == nil {
		return ctx, false
	}

	// Run error middleware
	ctx = mw.Run("error", ctx)
	if ctx.Error != nil {
		return ctx, true
	}

	return ctx, false
}
