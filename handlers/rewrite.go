package handlers

import (
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	errorForbidden = StatusError{403, errors.New("bad dscv")}
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Status returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

// Env represents options always present in handlers.
type Env struct {
	Proxy        *httputil.ReverseProxy
	Log          *logrus.Logger
	CustomHeader string
}

// Handler is a wrapper to satisfy http.Handler and to pass around an *Env context.
type Handler struct {
	*Env
	H func(e *Env, w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			h.Env.Log.Printf("HTTP %d - %s", e.Status(), e)

			http.Error(w, e.Error(), e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}

// Status is an http hangler used as a health/readiness check in k8s and openshift.
func Status(env *Env, w http.ResponseWriter, r *http.Request) error {
	_, err := w.Write([]byte("OK\n"))
	if err != nil {
		return err
	}
	return nil

}

// ProxyHandler sends http requests to upstream.
func ProxyHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	if env.CustomHeader != "" {
		values := strings.Split(env.CustomHeader, ":")
		r.Header.Add(values[0], values[1])
	}
	env.Log.WithFields(logrus.Fields{"path": r.URL}).Info("Forwarding url to upstream.")
	env.Proxy.ServeHTTP(w, r)
	return nil
}
