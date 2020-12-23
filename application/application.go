package application

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/carbocation/interpose"
	"github.com/gomodule/redigo/redis"
	gorilla_mux "github.com/gorilla/mux"
	"github.com/jfardello/tlsrproxy/handlers"
	"github.com/jfardello/tlsrproxy/internal/bytereplacer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var lock = &sync.Mutex{}

type responseHeadersTransport struct {
	oldnew []string
}

var (
	bodyReplacer    *bytereplacer.Replacer
	headersReplacer *strings.Replacer
)

//NewBodyReplacer return a body singleton replacer.
func NewBodyReplacer(oldnew ...string) *bytereplacer.Replacer {
	lock.Lock()
	defer lock.Unlock()
	if bodyReplacer == nil {
		bodyReplacer = bytereplacer.New(oldnew...)
	}
	return bodyReplacer
}

//NewHeaderReplacer return a body singleton replacer.
func NewHeaderReplacer(oldnew ...string) *strings.Replacer {
	lock.Lock()
	defer lock.Unlock()
	if headersReplacer == nil {
		headersReplacer = strings.NewReplacer(oldnew...)
	}
	return headersReplacer
}

//RoudTrip is used to mangle headers.
func (t responseHeadersTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	hr := NewHeaderReplacer(t.oldnew...)
	for key := range resp.Header {
		resp.Header.Set(key, hr.Replace(resp.Header.Get(key)))
	}

	return resp, nil
}

// New is the constructor for Application struct.
func New(config *viper.Viper) (*Application, error) {
	app := &Application{}
	app.config = config
	return app, nil
}

// Application is the application object that runs HTTP server.
type Application struct {
	config *viper.Viper
}

//MiddlewareStruct helps embed stuff into the real handlers.
func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.UseHandler(app.mux())
	return middle, nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

//UpdateResponse replaces the body pf a request with a modifyed one, golang cannot modify inplace the body.
func UpdateResponse(r *http.Response) error {
	b, _ := ioutil.ReadAll(r.Body)
	br := NewBodyReplacer("foo", "FoOo")
	replace := br.Replace(b)
	buf := bytes.NewBuffer(replace)
	r.Body = ioutil.NopCloser(buf)
	r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
	return nil
}

//NewProxy returns a configured httputil.ReverseProxy
func NewProxy(u *url.URL) *httputil.ReverseProxy {
	targetQuery := u.RawQuery
	RemoveHeaders := responseHeadersTransport{oldnew: []string{"header1", "header2"}}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Host = u.Host
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = singleJoiningSlash(u.Path, req.URL.Path)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery

			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
		},
		Transport:      RemoveHeaders,
		ModifyResponse: UpdateResponse,
	}
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

func (app *Application) mux() *gorilla_mux.Router {

	router := gorilla_mux.NewRouter()
	u, err := url.Parse(app.config.GetString("proxy.upstream"))

	env := handlers.Env{
		Log: &logrus.Logger{
			Out:   os.Stderr,
			Level: logrus.InfoLevel,
			Formatter: &logrus.TextFormatter{
				FullTimestamp: true,
			},
		},
	}
	if err == nil {
		env.Proxy = NewProxy(u)
	}

	router.Handle("/_health/status", handlers.Handler{Env: &env, H: handlers.Status})
	if env.Proxy != nil {
		router.PathPrefix("/").Handler(handlers.Handler{Env: &env, H: handlers.ProxyHandler})
	}

	return router
}
