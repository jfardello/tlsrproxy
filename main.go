package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tylerb/graceful"

	"github.com/jfardello/tlsrproxy/application"
)

func newConfig() (*viper.Viper, error) {
	c := viper.New()
	c.SetDefault("http_addr", ":8888")
	c.SetDefault("upstream", "http://localhost:8080")
	c.SetDefault("http_cert_file", "")
	c.SetDefault("http_key_file", "")
	c.SetDefault("http_drain_interval", "1s")
	c.SetDefault("custom_header", nil)

	c.AutomaticEnv()
	return c, nil
}

func originValidator(origin string) bool {
	config, err := newConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	for _, b := range strings.Split(config.GetString("cors_origins_allowed"), ",") {
		if b == origin {
			return true
		}
	}
	return false
}

func main() {
	config, err := newConfig()
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

	serverAddress := config.Get("http_addr").(string)

	certFile := config.Get("http_cert_file").(string)
	keyFile := config.Get("http_key_file").(string)
	drainIntervalString := config.Get("http_drain_interval").(string)

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
