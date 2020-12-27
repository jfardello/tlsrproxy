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
