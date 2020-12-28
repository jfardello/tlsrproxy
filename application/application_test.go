package application

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jfardello/tlsrproxy/internal/config"
)

var recordedHeader string = ""

func TestApplication(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Content-Type", "text/html")
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()
	c, _ := config.GetConf()
	c.Proxy.Upstream = ts.URL
	c.Proxy.Replaces.Response.Body = [][]string{
		{"Hello", "Hola"},
	}
	app, _ := New(c)
	middle, _ := app.MiddlewareStruct()
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	middle.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Error("Bad status")
	}
	if !strings.HasPrefix(string(body), "Hola, client") {
		t.Errorf("Bad replace. (%s)", string(body))
	}
}
func TestApplicationHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		recordedHeader = r.Header.Get("X-Foo")
		header := w.Header()
		header.Set("Content-Type", "text/html")
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()
	c, _ := config.GetConf()
	c.Proxy.Upstream = ts.URL
	c.Proxy.Replaces.Response.Body = nil
	c.Proxy.Replaces.Request.Headers = [][]string{{"wolololo", "walalala"}}
	app, _ := New(c)
	middle, _ := app.MiddlewareStruct()
	req := httptest.NewRequest("GET", "http://lol.com/bar", nil)
	req.Header.Set("X-Foo", "wolololo")
	w := httptest.NewRecorder()

	middle.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Error("Bad status")
	}
	if recordedHeader != "walalala" {
		t.Errorf("Bad header replacement: (%s)", recordedHeader)
	}
}
func TestDontReplacePDF(t *testing.T) {
	var text string = "fake text pdf"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := w.Header()
		header.Set("Content-Type", "application/pdf")
		fmt.Fprintln(w, text)
	}))
	defer ts.Close()
	c, _ := config.GetConf()
	c.Proxy.Upstream = ts.URL
	c.Proxy.Replaces.Response.Body = [][]string{{"fake", "error"}}
	app, _ := New(c)
	middle, _ := app.MiddlewareStruct()
	req := httptest.NewRequest("GET", "http://lol.com/pdf", nil)
	w := httptest.NewRecorder()

	middle.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Error("Bad status")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	//this content type shudn't be replaced as per the current config.
	if strings.HasPrefix(string(body), "error") {
		t.Errorf("Bad replace. (%s)", string(body))
	}
}

func TestUpstremaContext(t *testing.T) {
	ResetReplacers()
	var text string = "fake text"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := w.Header()
		header.Set("Content-Type", "text/html")
		fmt.Fprintln(w, text)
	}))
	defer ts.Close()
	c := &config.Conf{}
	c.Proxy.Upstream = ts.URL + "/pepe?foo=22"
	c.Proxy.Replaces.Response.Body = [][]string{{"fake", "changed"}}
	c.Proxy.Replaces.Response.Mimes = []string{"text/html"}
	config.SetConf(c)
	app, _ := New(c)
	middle, _ := app.MiddlewareStruct()
	req := httptest.NewRequest("GET", "http://lol.com/ninonino", nil)
	w := httptest.NewRecorder()

	middle.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Error("Bad status")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !strings.HasPrefix(string(body), "changed") {
		t.Errorf("Bad replace. (%s)", string(body))
	}
}

func TestUpstremaContext2(t *testing.T) {
	ResetReplacers()
	var text string = "fake text"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := w.Header()
		header.Set("Content-Type", "text/html")
		fmt.Fprintln(w, text)
	}))
	defer ts.Close()
	c := &config.Conf{}
	c.Proxy.Upstream = ts.URL + "/pepe?foo=22"
	c.Proxy.Replaces.Response.Body = [][]string{{"fake", "changed"}}
	c.Proxy.Replaces.Response.Mimes = []string{"text/html"}
	config.SetConf(c)
	app, _ := New(c)
	middle, _ := app.MiddlewareStruct()
	req := httptest.NewRequest("GET", "http://lol.com/ninonino?fofo=22", nil)
	w := httptest.NewRecorder()

	middle.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Error("Bad status")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !strings.HasPrefix(string(body), "changed") {
		t.Errorf("Bad replace. (%s)", string(body))
	}
}
func TestUpstremaDown(t *testing.T) {
	ResetReplacers()
	var text string = "fake text"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := w.Header()
		header.Set("Content-Type", "text/html")
		fmt.Fprintln(w, text)
	}))
	c := &config.Conf{}
	c.Proxy.Upstream = ts.URL
	c.Proxy.Replaces.Response.Body = [][]string{{"fake", "changed"}}
	c.Proxy.Replaces.Response.Mimes = []string{"text/html"}
	ts.Close()
	config.SetConf(c)
	app, _ := New(c)
	middle, _ := app.MiddlewareStruct()
	req := httptest.NewRequest("GET", "http://lol.com/ninonino?fofo=22", nil)
	w := httptest.NewRecorder()

	middle.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 502 {
		t.Errorf("Bad status (%d)", resp.StatusCode)
	}
}

func TestNoContentType(t *testing.T) {
	ResetReplacers()
	var text string = "some text"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		//If we just delete the header, go http machinery will add a default, so better set it to empty string.
		header.Set("Content-Type", "")
		fmt.Fprintln(w, text)
	}))
	defer ts.Close()
	c := &config.Conf{}
	c.Proxy.Upstream = ts.URL
	c.Proxy.Replaces.Response.Body = [][]string{{"some", "changed"}}
	c.Proxy.Replaces.Response.Mimes = []string{"text/html"}
	config.SetConf(c)
	app, _ := New(c)
	middle, _ := app.MiddlewareStruct()
	req := httptest.NewRequest("GET", "http://lol.com/ninonino?fofo=22", nil)
	w := httptest.NewRecorder()

	middle.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Error("Bad status")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if strings.HasPrefix(string(body), "changed") {
		t.Errorf("Bad replace. (%#v)", string(body))
	}
}
