package config

import (
	"bytes"
	"errors"
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// protocol is used to determine the http-handler to be used for an url.
type protocol int

const (
	_ protocol = iota
	NSQ
	HTTP
)

func (p protocol) String() string {
	return [...]string{"_", "NSQ", "HTTP"}[p]
}

type NSQConf struct {
	Topic    string   `yaml:"topic"`
	TCPAddrs []string `yaml:"dest_tcp_addr"`
}

type HTTPConf struct {
	Host string `yaml:"host"`
}

// Does not support query params.
type URL struct {
	Path   string   `yaml:"path"`
	Method string   `yaml:"method"`
	NSQ    NSQConf  `yaml:"nsq"`
	HTTP   HTTPConf `yaml:"http"`
}

func (u URL) Protocol() (protocol, error) {
	if u.NSQ.Topic != "" {
		return NSQ, nil
	} else if u.HTTP.Host != "" {
		return HTTP, nil
	}
	return 0, errors.New("URL is missing protocol")
}

// config represents a server configuration read from a YAML file.
type Config struct {
	URLs []URL `yaml:"urls"`
}

// FromFile loads a configuration from file.
func FromFile(cfgpath string) (*Config, error) {
	f, err := os.Open(cfgpath)
	if err != nil {
		return nil, err
	}
	return load(f)
}

// load loads configuration from an io.Reader.
// Note, the configuration does not get validated or sanitized.
func load(in io.Reader) (*Config, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(in)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	err = yaml.Unmarshal(buf.Bytes(), c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
