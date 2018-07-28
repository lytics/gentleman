package gentleman

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/plugins/multipart"
	"github.com/lytics/gentleman/utils"
)

var (
	requestDump = `GET / HTTP/1.1
Host: %s
Content-Type: application/json
User-Agent: gentleman/2.0.3

{"foo": "bar", "id": 123}`
)

func TestRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest()
	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		h.Next(ctx)
	})

	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		u, _ := url.Parse(ts.URL)
		ctx.Request.URL = u
		h.Next(ctx)
	})

	res, err := req.Send()
	utils.Equal(t, err, nil)
	utils.NotEqual(t, res.RawRequest.URL, nil)
	utils.Equal(t, res.StatusCode, 200)
}

func TestRequestString(t *testing.T) {
	body := `{"foo": "bar", "id": 123}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		utils.Equal(t, err, nil)
		utils.Equal(t, fmt.Sprintf("%s", b), body)

		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest().URL(ts.URL).JSON(body)
	utils.Equal(t, req.String(), body)
	_, err := req.Send()
	utils.Equal(t, err, nil)
}

func TestRequestDump(t *testing.T) {
	body := `{"foo": "bar", "id": 123}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		utils.Equal(t, err, nil)
		utils.Equal(t, fmt.Sprintf("%s", b), body)

		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	d := fmt.Sprintf(requestDump, ts.Listener.Addr().String())

	req := NewRequest().URL(ts.URL).JSON(body)
	dump, err := req.Dump()
	utils.Equal(t, err, nil)
	//go adds \r with each new lines apparently
	dump = strings.Replace(dump, "\r", "", -1)

	_, err = req.Send()
	utils.Equal(t, err, nil)
	utils.Equal(t, dump, d)
}

func TestRequestAlreadyDispatched(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest()
	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		h.Next(ctx)
	})

	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		u, _ := url.Parse(ts.URL)
		ctx.Request.URL = u
		h.Next(ctx)
	})

	res, err := req.Send()
	utils.Equal(t, err, nil)
	utils.NotEqual(t, res.RawRequest.URL, nil)
	utils.Equal(t, res.StatusCode, 200)

	res, err = req.Send()
	utils.NotEqual(t, err, nil)
}

func TestMiddlewareErrorInjectionAndInterception(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest()
	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		h.Next(ctx)
	})

	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		u, _ := url.Parse(ts.URL)
		ctx.Request.URL = u
		h.Error(ctx, errors.New("Error"))
	})

	req.UseError(func(ctx *context.Context, h context.Handler) {
		ctx.Error = nil
		h.Next(ctx)
	})

	res, err := req.Send()
	utils.Equal(t, err, nil)
	utils.NotEqual(t, res.RawRequest.URL, nil)
	utils.Equal(t, res.StatusCode, 200)
}

func TestRequestResponseMiddleware(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest()
	req.URL(ts.URL)
	req.UseRequest(func(c *context.Context, h context.Handler) {
		c.Request.Header.Set("Client", "go")
		h.Next(c)
	})
	req.UseResponse(func(c *context.Context, h context.Handler) {
		c.Response.Header.Set("Server", c.Request.Header.Get("Client"))
		h.Next(c)
	})

	res, err := req.Do()
	utils.Equal(t, err, nil)
	utils.Equal(t, res.StatusCode, 200)
	utils.Equal(t, res.Header.Get("Server"), "go")
}

func TestRequestCustomPhaseMiddleware(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest()
	req.URL(ts.URL)
	req.UseHandler("request", func(c *context.Context, h context.Handler) {
		c.Request.Header.Set("Client", "go")
		h.Next(c)
	})
	req.UseHandler("response", func(c *context.Context, h context.Handler) {
		c.Response.Header.Set("Server", c.Request.Header.Get("Client"))
		h.Next(c)
	})

	res, err := req.Do()
	utils.Equal(t, err, nil)
	utils.Equal(t, res.StatusCode, 200)
	utils.Equal(t, res.Header.Get("Server"), "go")
}

func TestRequestOverwriteTargetURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest().URL("http://invalid")
	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.URL, _ = url.Parse(ts.URL)
		h.Next(ctx)
	})

	res, err := req.Do()
	utils.Equal(t, err, nil)
	utils.Equal(t, res.StatusCode, 200)
}

func TestRequestMux(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest()
	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		u, _ := url.Parse(ts.URL)
		ctx.Request.URL = u
		h.Next(ctx)
	})

	req.Mux().AddMatcher(func(ctx *context.Context) bool {
		return ctx.GetString("$phase") == "request" && ctx.Request.Method == "GET"
	}).UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.Header.Set("mux", "true")
		h.Next(ctx)
	})

	res, err := req.Send()
	utils.Equal(t, err, nil)
	utils.Equal(t, res.StatusCode, 200)
	utils.Equal(t, res.RawRequest.Header.Get("mux"), "true")
}

func TestRequestInterceptor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Server should not be reached!")
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest()
	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.Header.Set("Client", "gentleman")

		ctx.Response.StatusCode = 201
		ctx.Response.Status = "201 Created"

		ctx.Response.Header.Set("Server", "gentleman")
		utils.WriteBodyString(ctx.Response, "Hello, gentleman")

		h.Stop(ctx)
	})
	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		t.Fatal("middleware should not be called")
		h.Next(ctx)
	})

	res, err := req.Do()
	utils.Equal(t, err, nil)
	utils.Equal(t, res.StatusCode, 201)
	utils.Equal(t, res.RawRequest.Header.Get("Client"), "gentleman")
	utils.Equal(t, res.RawResponse.Header.Get("Server"), "gentleman")
	utils.Equal(t, res.String(), "Hello, gentleman")
}

func TestRequestTimeout(t *testing.T) {
	if runtime.Version() != "go1.6" {
		return
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1000 * time.Millisecond)
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest().URL(ts.URL)

	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Client.Timeout = 50 * time.Millisecond
		h.Next(ctx)
	})

	res, err := req.Send()
	utils.NotEqual(t, err, nil)
	utils.Equal(t, strings.Contains(err.Error(), "net/http: request canceled"), true)
	utils.Equal(t, res.StatusCode, 0)
}

func TestRequestCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	req := NewRequest().URL(ts.URL)

	req.UseRequest(func(ctx *context.Context, h context.Handler) {
		if tra, ok := ctx.Client.Transport.(*http.Transport); !ok {
			t.Fatal("transport does not implement CancelRequest(*http.Request)")
		} else {
			tra.CancelRequest(ctx.Request)
		}
		h.Stop(ctx)
	})

	res, err := req.Do()
	utils.Equal(t, err, nil)
	utils.Equal(t, res.StatusCode, 0)
}

func TestRequestGoroutines(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	times := 10
	var wg sync.WaitGroup

	for i := 0; i < times; i++ {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			res, err := NewRequest().URL(url).Send()
			utils.Equal(t, err, nil)
			utils.Equal(t, res.Ok, true)
			utils.Equal(t, res.StatusCode, 200)
		}(ts.URL)
	}

	wg.Wait()
}

// Test API methods

func TestRequestMethod(t *testing.T) {
	req := NewRequest()
	req.Method("POST")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Method, "POST")
}

func TestRequestURL(t *testing.T) {
	url := "http://foo.com"
	req := NewRequest()
	req.URL(url)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.String(), url)
}

func TestRequestBaseURL(t *testing.T) {
	url := "http://foo.com/bar/baz"
	req := NewRequest()
	req.BaseURL(url)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.String(), "http://foo.com")
}

func TestRequestPath(t *testing.T) {
	url := "http://foo.com/bar/baz"
	path := "/foo/baz"
	req := NewRequest()
	req.URL(url)
	req.Path(path)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.String(), "http://foo.com/foo/baz")
}

func TestRequestAddPath(t *testing.T) {
	url := "http://foo.com/bar/baz"
	path := "/foo/baz"
	req := NewRequest()
	req.URL(url)
	req.AddPath(path)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.String(), "http://foo.com/bar/baz/foo/baz")
}

func TestRequestPathParam(t *testing.T) {
	url := "http://foo.com/bar/baz"
	path := "/:foo/bar/:baz"
	req := NewRequest()
	req.URL(url)
	req.Path(path)
	req.Param("foo", "baz")
	req.Param("baz", "foo")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.String(), "http://foo.com/baz/bar/foo")
}

func TestRequestPathParams(t *testing.T) {
	url := "http://foo.com/bar/baz"
	path := "/:foo/bar/:baz"
	req := NewRequest()
	req.URL(url)
	req.Path(path)
	req.Params(map[string]string{"foo": "baz", "baz": "foo"})
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.String(), "http://foo.com/baz/bar/foo")
}

func TestRequestSetQuery(t *testing.T) {
	req := NewRequest()
	req.SetQuery("foo", "bar")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.RawQuery, "foo=bar")
}

func TestRequestAddQuery(t *testing.T) {
	req := NewRequest()
	req.AddQuery("foo", "bar")
	req.AddQuery("foo", "bar")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.RawQuery, "foo=bar&foo=bar")
}

func TestRequestSetQueryParams(t *testing.T) {
	req := NewRequest()
	param := url.Values{}
	param.Set("foo", "bar")
	req.SetQueryParams(param)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.URL.RawQuery, "foo=bar")
}

func TestRequestSetHeader(t *testing.T) {
	req := NewRequest()
	req.SetHeader("foo", "bar")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Header.Get("foo"), "bar")
}

func TestRequestAddHeader(t *testing.T) {
	req := NewRequest()
	req.AddHeader("foo", "baz")
	req.AddHeader("foo", "bar")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Header.Get("foo"), "baz")
}

func TestRequestSetHeaders(t *testing.T) {
	req := NewRequest()
	req.SetHeaders(map[string]string{"foo": "baz", "baz": "foo"})
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Header.Get("foo"), "baz")
	utils.Equal(t, req.Context.Request.Header.Get("baz"), "foo")
}

func TestRequestDelHeaders(t *testing.T) {
	req := NewRequest()
	req.SetHeaders(map[string]string{"foo": "baz", "baz": "foo"})
	req.DelHeader("baz")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Header.Get("foo"), "baz")
	utils.Equal(t, req.Context.Request.Header.Get("baz"), "")
}

func TestRequestAddCookie(t *testing.T) {
	req := NewRequest()
	cookie := &http.Cookie{Name: "foo", Value: "bar"}
	req.AddCookie(cookie)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Header.Get("Cookie"), "foo=bar")
}

func TestRequestAddCookies(t *testing.T) {
	req := NewRequest()
	cookies := []*http.Cookie{{Name: "foo", Value: "bar"}}
	req.AddCookies(cookies)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Header.Get("Cookie"), "foo=bar")
}

// func TestRequestCookieJar(t *testing.T) {
// 	req := NewRequest()
// 	req.CookieJar()
// 	req.Middleware.Run("request", req.Context)
// 	utils.NotEqual(t, req.Context.Client.Jar, nil)
// }

func TestRequestType(t *testing.T) {
	req := NewRequest()
	req.Type("json")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, req.Context.Request.Header.Get("Content-Type"), "application/json")
}

func TestRequestBody(t *testing.T) {
	reader := bytes.NewReader([]byte("foo bar"))
	req := NewRequest()
	req.Body(reader)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, int(req.Context.Request.ContentLength), 7)
	utils.Equal(t, req.Context.Request.Header.Get("Content-Type"), "")
	body, _ := ioutil.ReadAll(req.Context.Request.Body)
	utils.Equal(t, string(body), "foo bar")
}

func TestRequestBodyString(t *testing.T) {
	req := NewRequest()
	req.BodyString("foo bar")
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, int(req.Context.Request.ContentLength), 7)
	utils.Equal(t, req.Context.Request.Header.Get("Content-Type"), "")
	body, _ := ioutil.ReadAll(req.Context.Request.Body)
	utils.Equal(t, string(body), "foo bar")
}

func TestRequestJSON(t *testing.T) {
	req := NewRequest()
	req.JSON(map[string]string{"foo": "bar"})
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, int(req.Context.Request.ContentLength), 14)
	utils.Equal(t, req.Context.Request.Header.Get("Content-Type"), "application/json")
	body, _ := ioutil.ReadAll(req.Context.Request.Body)
	utils.Equal(t, string(body)[:13], `{"foo":"bar"}`)
}

func TestRequestXML(t *testing.T) {
	type xmlTest struct {
		Name string `xml:"name>first"`
	}
	xml := xmlTest{Name: "foo"}

	req := NewRequest()
	req.XML(xml)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, int(req.Context.Request.ContentLength), 50)
	utils.Equal(t, req.Context.Request.Header.Get("Content-Type"), "application/xml")
	body, _ := ioutil.ReadAll(req.Context.Request.Body)
	utils.Equal(t, string(body), `<xmlTest><name><first>foo</first></name></xmlTest>`)
}

func TestRequestForm(t *testing.T) {
	reader := bytes.NewReader([]byte("hello world"))
	fields := map[string]multipart.Values{
		"foo": {"data=bar"},
		"bar": {"data=baz"},
	}
	data := multipart.FormData{
		Files: []multipart.FormFile{{Name: "foo", Reader: reader}},
		Data:  fields,
	}

	req := NewRequest()
	req.Form(data)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, strings.Contains(req.Context.Request.Header.Get("Content-Type"), "multipart/form-data"), true)
	body, _ := ioutil.ReadAll(req.Context.Request.Body)
	utils.Equal(t, strings.Contains(string(body), "data=bar"), true)
	utils.Equal(t, strings.Contains(string(body), "data=baz"), true)
}

func TestRequestFile(t *testing.T) {
	reader := bytes.NewReader([]byte("hello world"))
	req := NewRequest()
	req.File("foo", reader)
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, strings.Contains(req.Context.Request.Header.Get("Content-Type"), "multipart/form-data"), true)
	body, _ := ioutil.ReadAll(req.Context.Request.Body)
	utils.Equal(t, strings.Contains(string(body), "hello world"), true)
}

func TestRequestFiles(t *testing.T) {
	reader1 := bytes.NewReader([]byte("content1"))
	reader2 := bytes.NewReader([]byte("content2"))
	file1 := multipart.FormFile{Name: "file1.txt", Reader: reader1}
	file2 := multipart.FormFile{Name: "file2.txt", Reader: reader2}

	req := NewRequest()
	req.Files([]multipart.FormFile{file1, file2})
	req.Middleware.Run("request", req.Context)
	utils.Equal(t, strings.Contains(req.Context.Request.Header.Get("Content-Type"), "multipart/form-data"), true)

	body, _ := ioutil.ReadAll(req.Context.Request.Body)
	utils.Equal(t, strings.Contains(string(body), "content1"), true)
	utils.Equal(t, strings.Contains(string(body), "content2"), true)
}

func TestRequestClone(t *testing.T) {
	req1 := NewRequest()
	req1.UseRequest(func(c *context.Context, h context.Handler) {})
	req1.Context.Set("foo", "bar")
	req2 := req1.Clone()
	utils.Equal(t, req1 != req2, true)
	utils.Equal(t, req2.Context.GetString("foo"), req1.Context.GetString("foo"))
	utils.Equal(t, len(req2.Middleware.GetStack()), 1)
}

func BenchmarkSimpleRequestGet(b *testing.B) {
	ts := createEchoServer()
	defer ts.Close()

	for n := 0; n < b.N; n++ {
		NewRequest().URL(ts.URL).Send()
	}
}

func BenchmarkSimpleRequestPostSmallString(b *testing.B) {
	body := randomString(200)
	ts := createEchoServer()
	defer ts.Close()

	for n := 0; n < b.N; n++ {
		NewRequest().URL(ts.URL).BodyString(body).Send()
	}
}

func BenchmarkSimpleRequestPostLargeString(b *testing.B) {
	body := randomString(10000)
	ts := createEchoServer()
	defer ts.Close()

	for n := 0; n < b.N; n++ {
		NewRequest().URL(ts.URL).BodyString(body).Send()
	}
}

func BenchmarkRequestPlugins(b *testing.B) {
	ts := createEchoServer()
	defer ts.Close()

	for n := 0; n < b.N; n++ {
		req := NewRequest().URL(ts.URL)
		for i := 0; i < 10; i++ {
			req.UseRequest(headerRequestMiddleware)
			req.UseResponse(headerResponseMiddleware)
		}
		req.Send()
	}
}

func headerRequestMiddleware(ctx *context.Context, h context.Handler) {
	ctx.Request.Header.Set("foo", ctx.Request.Header.Get("foo")+"bar")
	h.Next(ctx)
}

func headerResponseMiddleware(ctx *context.Context, h context.Handler) {
	ctx.Response.Header.Set("foo", ctx.Response.Header.Get("foo")+"bar")
	h.Next(ctx)
}

func createEchoServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
}

func randomString(n int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
