package Golf

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func makeTestHTTPRequest(body io.Reader, method, url string) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil
	}
	return req
}

func makeTestContext(method, url string) *Context {
	r := makeTestHTTPRequest(nil, method, url)
	w := httptest.NewRecorder()
	app := New()
	return NewContext(r, w, app)
}

func TestContextCreate(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/foo/bar/")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	if ctx == nil {
		t.Errorf("Can not create context.")
	}
}

func TestCookieSet(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/foo/bar/")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	ctx.SetCookie("foo", "bar", 0)
	ctx.Send()
	if w.HeaderMap.Get("Set-Cookie") != `foo=bar; Path=/` {
		t.Errorf("Cookie test failed: %q != %q", w.HeaderMap.Get("Set-Cookie"), `foo=bar; Path=/`)
	}
}

func TestCookieSetWithExpire(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/foo/bar/")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	ctx.SetCookie("foo", "bar", 3600)
	ctx.Send()
	rawCookie := w.HeaderMap.Get("Set-Cookie")
	rawRequest := fmt.Sprintf("GET / HTTP/1.0\r\nCookie: %s\r\n\r\n", rawCookie)
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(rawRequest)))
	if err == nil {
		cookies := req.Cookies()
		cookie := cookies[3]
		if cookie.Value != "3600" {
			t.Errorf("Can not set cookie with expiration correctly.")
		}
	}
}

func TestTemplateLoader(t *testing.T) {
	ctx := makeTestContext("GET", "/")
	ctx.Loader("admin")
	if ctx.templateLoader != "admin" {
		t.Errorf("Can not set templateLoader for Context.")
	}
}

func TestQuery(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/search?q=foo&p=bar")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	q, err := ctx.Query("q")
	if err != nil {
		t.Errorf("Can not retrieve a query.")
	} else {
		if q != "foo" {
			t.Errorf("Can not retrieve the correct query `q`.")
		}
	}
	p, err := ctx.Query("p")
	if err != nil {
		t.Errorf("Can not retrieve a query.")
	} else {
		if p != "bar" {
			t.Errorf("Can not retrieve the correct query `p`.")
		}
	}
}

func TestQueries(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/search?myarray=value1&myarray=value2&myarray=value3")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	q, err := ctx.Query("myarray", 2)
	if err != nil {
		t.Errorf("Can not retrieve a query.")
	}
	if q != "value3" {
		t.Errorf("Can not correctly retrive a query.")
	}
}

func TestQueryNotFound(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/search?myarray=value1&myarray=value2&myarray=value3")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	q, err := ctx.Query("query")
	if err == nil || q != "" {
		t.Errorf("Can not raise error when query not found.")
	}
}

func makeNewContext(method, url string) *Context {
	r := makeTestHTTPRequest(nil, method, url)
	w := httptest.NewRecorder()
	app := New()
	return NewContext(r, w, app)
}

func TestRedirection(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	ctx.Redirect("/foo")
	ctx.Send()
	if w.HeaderMap.Get("Location") != `/foo` {
		t.Errorf("Can not perform a 301 redirection.")
	}
}

func TestWrite(t *testing.T) {
	ctx := makeNewContext("GET", "/foo")
	ctx.Write("hello world")
	if !reflect.DeepEqual(ctx.Body, []byte("hello world")) {
		t.Errorf("Context.Write failed.")
	}
}

func TestAbort(t *testing.T) {
	r := makeTestHTTPRequest(nil, "GET", "/")
	w := httptest.NewRecorder()
	app := New()
	ctx := NewContext(r, w, app)
	ctx.Abort(500)
	if w.Code != 500 || !ctx.IsSent {
		t.Errorf("Can not abort a context.")
	}
}

func TestRenderFromString(t *testing.T) {
	cases := []struct {
		src    string
		args   map[string]interface{}
		output string
	}{
		{
			"foo {{.Title}} bar",
			map[string]interface{}{"Title": "Hello World"},
			"foo Hello World bar",
		},
	}

	for _, c := range cases {
		r := makeTestHTTPRequest(nil, "GET", "/")
		w := httptest.NewRecorder()
		app := New()
		ctx := NewContext(r, w, app)
		ctx.RenderFromString(c.src, c.args)
		ctx.Send()
		if w.Body.String() != c.output {
			t.Errorf("Can not render from string correctly: %v != %v", w.Body.String(), c.output)
		}
	}
}

func TestJSON(t *testing.T) {
	cases := []struct {
		input  map[string]interface{}
		output string
	}{
		{
			map[string]interface{}{"status": "success", "code": 200},
			`{"code":200,"status":"success"}`,
		},
	}

	for _, c := range cases {
		r := makeTestHTTPRequest(nil, "GET", "/")
		w := httptest.NewRecorder()
		app := New()
		ctx := NewContext(r, w, app)
		ctx.JSON(c.input)
		ctx.Send()
		if w.Body.String() != c.output {
			t.Errorf("Can not return JSON correctly: %v != %v", w.Body.String(), c.output)
		}
		if w.HeaderMap.Get("Content-Type") != `application/json` {
			t.Errorf("Content-Type didn't set properly when calling Context.JSON.")
		}
	}
}
