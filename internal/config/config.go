package config

import (
	"errors"
	"strings"
	"sync"

	"github.com/jfardello/tlsrproxy/libhttp"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

//ConfigFileName is the default config for the rewriting proxy.
var ConfigFileName string = "tlsrproxy"
var conf *Conf
var lock = &sync.Mutex{}

func init() {
	var err error
	conf, err = newConfig()
	if err != nil {
		logrus.Info("No config file found, using defaults.")
	}
}

//GetConf returns the global vper object
func GetConf() (*Conf, error) {
	if conf == nil {
		return nil, errors.New("Can't get config")
	}
	return conf, nil
}

//SetConf is used mainly in unit tests.
func SetConf(c *Conf) error {
	lock.Lock()
	defer lock.Unlock()
	conf = c
	return nil
}

type Server struct {
	HTTPAddr string `mapstructure:"http_addr"`
	Cert     string `mapstructure:"cert"`
	Key      string `mapstructure:"key"`
	Drain    string `mapstructure:"drain"`
}

type Proxy struct {
	Upstream string   `mapstructure:"upstream"`
	Mimes    []string `mapstructure:"mimes"`
	Replaces Replaces
}

type Replaces struct {
	Request Request
	Response
}

type Request struct {
	Headers Headers
}

type Response struct {
	Headers Headers  `mapstructure:"headers"`
	Body    Body     `mapstructure:"body"`
	Mimes   []string `mapstructure:"mimes"`
}

type Headers PairList
type Body PairList

type PairList [][]string

func (h *Headers) Flatttern() []string {
	return libhttp.ToSlice(*h)
}

func (b *Body) Flattern() []string {
	return libhttp.ToSlice(*b)
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
	c.SetDefault("proxy.replaces.response.body", [][]string{
		{"http://", "https://"},
		{"foo", "bar"},
	})
	c.SetDefault("proxy.replaces.response.headers", [][]string{
		{"bart", "bert"},
		{"foo", "bar"},
	})
	c.SetDefault("proxy.replaces.request.headers", [][]string{{"wolololo", "walalala"}})
	c.SetDefault("proxy.replaces.response.mimes", []string{"text/html"})
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
