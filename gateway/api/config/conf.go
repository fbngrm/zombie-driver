package config

import (
	"bytes"
	"errors"
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type protocol int

const (
	_ protocol = iota
	NSQ
	HTTP
)

func (p protocol) String() string {
	return [...]string{"_", "NSQ", "HTTP"}[p]
}

type URL struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
	NSQ    struct {
		Topic    string   `yaml:"topic"`
		TCPAddrs []string `yaml:"dest_tcp_addr"`
	} `yaml:"nsq"`
	HTTP struct {
		Host string `yaml:"host"`
	} `yaml:"http"`
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
// NOTE: The configuration does not get validated or sanitized.
type Config struct {
	URLs []URL `yaml:"urls"`
}

func FromFile(cfgpath string) (*Config, error) {
	f, err := os.Open(cfgpath)
	if err != nil {
		return nil, err
	}
	return load(f)
}

// load loads configuration from an io.Reader.
// Note: the configuration is not checked for validity.
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
