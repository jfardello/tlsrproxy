package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

//ConfigFileName is the default config for the rewriting proxy.
var ConfigFileName string = "tlsrproxy"
var conf *Conf

func init() {
	var err error
	conf, err = newConfig()
	if err != nil {
		panic(err)
	}
}

//GetConf returns the global vper object
func GetConf() (*Conf, error) {
	if conf == nil {
		return nil, errors.New("Can't get config")
	}
	return conf, nil
}

type Server struct {
	HTTPAddr string `mapstructure:"http_addr"`
	Cert     string `mapstructure:"cert"`
	Key      string `mapstructure:"key"`
	Drain    string `mapstructure:"drain"`
}

type Proxy struct {
	Upstream               string     `mapstructure:"upstream"`
	BodyReplaces           [][]string `mapstructure:"bodyreplaces"`
	HeadersReplaces        [][]string `mapstructure:"headersreplaces"`
	HeadersRequestReplaces [][]string `mapstructure:"headersreqreplaces"`
}

type Conf struct {
	Server Server
	Proxy  Proxy
}

func newConfig() (*Conf, error) {
	c := viper.New()
	c.SetDefault("server.http_addr", ":8888")
	c.SetDefault("proxy.upstream", "http://localhost:8080")
	c.SetDefault("server.cert", "")
	c.SetDefault("server.key", "")
	c.SetDefault("server.drain", "1s")
	c.SetDefault("proxy.bodyreplaces", [][]string{
		{"http://", "https://"},
		{"foo", "bar"},
	})
	c.SetDefault("proxy.headersreplaces", [][]string{
		{"http://", "https://"},
		{"foo", "bar"},
	})
	c.SetConfigName(ConfigFileName)
	c.AddConfigPath("/config")
	c.AddConfigPath(".")
	c.AutomaticEnv()
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := c.ReadInConfig()
	conf := &Conf{}
	c.Unmarshal(conf)

	return conf, err
}
