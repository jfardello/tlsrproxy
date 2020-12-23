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

//ConfigFileName is the default config for the rewriting proxy.
var ConfigFileName string = "tlsrproxy"

func newConfig() (*viper.Viper, error) {
	c := viper.New()
	c.SetDefault("server.http_addr", ":8888")
	c.SetDefault("proxy.upstream", "http://localhost:8080")
	c.SetDefault("server.cert", "")
	c.SetDefault("server.key", "")
	c.SetDefault("server.drain", "1s")
	c.SetDefault("proxy.bodyreplaces", [][]string{
		[]string{"http://", "https://"},
		[]string{"foo", "bar"},
	})
	c.SetDefault("proxy.headersreplaces", [][]string{
		[]string{"http://", "https://"},
		[]string{"foo", "bar"},
	})
	c.SetConfigName(ConfigFileName)
	c.AddConfigPath(".")
	c.AddConfigPath("/config")
	c.AutomaticEnv()
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := c.ReadInConfig()
	return c, err
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

	serverAddress := config.Get("server.http_addr").(string)

	certFile := config.Get("server.cert").(string)
	keyFile := config.Get("server.key").(string)
	drainIntervalString := config.Get("server.drain").(string)

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
	logrus.Infoln("Forwarding to upstream on " + config.GetString("proxy.upstream"))

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
