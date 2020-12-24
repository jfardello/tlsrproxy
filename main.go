package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tylerb/graceful"

	"github.com/jfardello/tlsrproxy/application"
	"github.com/jfardello/tlsrproxy/internal/config"
)

func main() {
	config, err := config.GetConf()
	if err != nil {
		logrus.Fatal(err)
	}

	app, err := application.New(config)
	if err != nil {
		logrus.Fatal(err)
	}

	middle, err := app.MiddlewareStruct()
	if err != nil {
		logrus.Fatal(err)
	}

	serverAddress := config.Server.HTTPAddr

	certFile := config.Server.Cert
	keyFile := config.Server.Key
	drainIntervalString := config.Server.Drain

	drainInterval, err := time.ParseDuration(drainIntervalString)
	if err != nil {
		logrus.Fatal(err)
	}

	srv := &graceful.Server{
		Timeout: drainInterval,
		Server: &http.Server{Addr: serverAddress,
			Handler: http.Handler(middle)},
	}

	logrus.Infoln("Running HTTP server on " + serverAddress)
	logrus.Infoln("Forwarding to upstream on " + config.Proxy.Upstream)

	if certFile != "" && keyFile != "" {
		fmt.Println("Serving with TLS enabled")
		err = srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		fmt.Println("Warning! Serving clear text http!")
		err = srv.ListenAndServe()
	}

	if err != nil {
		logrus.Fatal(err)
	}
}
