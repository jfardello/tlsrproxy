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
	"time"

	"github.com/carbocation/interpose"
	"github.com/gomodule/redigo/redis"
	gorilla_mux "github.com/gorilla/mux"
	"github.com/jfardello/tlsrproxy/handlers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type responseHeadersTransport struct {
	headers []string
}

func (t responseHeadersTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	for _, each := range t.headers {
		resp.Header.Del(each)
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

func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.UseHandler(app.mux())
	return middle, nil
}

func getHmacParam(r *http.Request) string {
	raw, _ := url.PathUnescape(r.URL.Query().Get("hmac"))
	return raw
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
	replace := bytes.ReplaceAll(b, []byte("http://"), []byte("https://"))
	buf := bytes.NewBuffer(replace)
	//buf.Write(b)
	r.Body = ioutil.NopCloser(buf)
	r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
	return nil
}

//NewProxy returns a configured httputil.ReverseProxy
func NewProxy(u *url.URL) *httputil.ReverseProxy {
	targetQuery := u.RawQuery
	RemoveHeaders := responseHeadersTransport{headers: []string{"Access-Control-Allow-Origin", "Access-Control-Allow-Credentials"}}

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
	u, err := url.Parse(app.config.GetString("upstream"))

	env := handlers.Env{
		CustomHeader: app.config.GetString("custom_header"),
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
