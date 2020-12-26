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
	"github.com/jfardello/tlsrproxy/internal/config"
	"github.com/jfardello/tlsrproxy/libhttp"
	"github.com/sirupsen/logrus"
)

var lock = &sync.Mutex{}

type responseHeadersTransport struct {
	oldnew        []string
	requestOldnew []string
}

var (
	bodyReplacer           *bytereplacer.Replacer = nil
	headersReplacer        *strings.Replacer      = nil
	headersRequestReplacer *strings.Replacer      = nil
)

//NewBodyReplacer return a body singleton replacer.
func NewBodyReplacer(oldnew []string) *bytereplacer.Replacer {
	lock.Lock()
	defer lock.Unlock()
	if bodyReplacer == nil {
		bodyReplacer = bytereplacer.New(oldnew...)
	}
	return bodyReplacer
}

//NewHeaderReplacer return a body singleton replacer.
func NewHeaderReplacer(oldnew []string) *strings.Replacer {
	lock.Lock()
	defer lock.Unlock()
	if headersReplacer == nil {
		headersReplacer = strings.NewReplacer(oldnew...)
	}
	return headersReplacer
}

//NewHeaderRequestReplacer return a body singleton replacer.
func NewHeaderRequestReplacer(oldnew []string) *strings.Replacer {
	lock.Lock()
	defer lock.Unlock()
	if headersRequestReplacer == nil {
		headersRequestReplacer = strings.NewReplacer(oldnew...)
	}
	return headersRequestReplacer
}

//RoudTrip is used to mangle headers.
func (t responseHeadersTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	//headers we sent
	reqRepl := NewHeaderRequestReplacer(t.requestOldnew)
	for key := range r.Header {
		r.Header.Set(key, reqRepl.Replace(r.Header.Get(key)))
	}
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	hr := NewHeaderReplacer(t.oldnew)
	//Headers we get back from upstream
	for key := range resp.Header {
		resp.Header.Set(key, hr.Replace(resp.Header.Get(key)))
	}
	return resp, nil
}

// New is the constructor for Application struct.
func New(config *config.Conf) (*Application, error) {
	app := &Application{}
	app.config = config
	return app, nil
}

// Application is the application object that runs HTTP server.
type Application struct {
	config *config.Conf
}

//MiddlewareStruct helps embed stuff into the real handlers.
func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.UseHandler(app.mux())
	return middle, nil
}

//UpdateResponse replaces the body pf a request with a modifyed one, golang cannot modify inplace the body.
func UpdateResponse(r *http.Response) error {
	c, _ := config.GetConf()
	if validateMime(r, c) == false {
		logrus.Info("HERE")
		return nil
	}
	b, _ := ioutil.ReadAll(r.Body)
	args, err := c.Proxy.Replaces.Response.Body.Flattern()
	if err != nil {
		logrus.Errorf("To slice failed: %s", err)
		return err
	}
	br := NewBodyReplacer(args)
	replace := br.Replace(b)
	buf := bytes.NewBuffer(replace)
	r.Body = ioutil.NopCloser(buf)
	r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
	return nil
}

func validateMime(r *http.Response, c *config.Conf) bool {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return false
	}
	return libhttp.Contains(c.Proxy.Replaces.Response.Mimes, ct)

}

//NewProxy returns a configured httputil.ReverseProxy
func NewProxy(u *url.URL) (*httputil.ReverseProxy, error) {
	c, _ := config.GetConf()
	targetQuery := u.RawQuery
	args, err := c.Proxy.Replaces.Response.Headers.Flatttern()
	if err != nil {
		return nil, err
	}
	ra, err := c.Proxy.Replaces.Request.Headers.Flatttern()
	transport := responseHeadersTransport{oldnew: args, requestOldnew: ra}
	if err != nil {
		return nil, err
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Host = u.Host
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = libhttp.SingleJoiningSlash(u.Path, req.URL.Path)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery

			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
		},
		Transport:      transport,
		ModifyResponse: UpdateResponse,
	}, nil
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
	u, err := url.Parse(app.config.Proxy.Upstream)

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
		env.Proxy, err = NewProxy(u)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	router.Handle("/_health/status", handlers.Handler{Env: &env, H: handlers.Status})
	if env.Proxy != nil {
		router.PathPrefix("/").Handler(handlers.Handler{Env: &env, H: handlers.ProxyHandler})
	}

	return router
}
