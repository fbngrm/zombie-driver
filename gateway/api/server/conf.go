package server

import (
	"errors"
	"io/ioutil"

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
		Topic string `yaml:"topic"`
	} `yaml:"nsq"`
	HTTP struct {
		Host string `yaml:"host"`
	} `yaml:"http"`
}

func (u URL) protocol() (protocol, error) {
	if u.NSQ.Topic != "" {
		return NSQ, nil
	} else if u.HTTP.Host != "" {
		return HTTP, nil
	}
	return 0, errors.New("missing protocol")
}

// config represents a server configuration read from a YAML file.
type config struct {
	URLs []URL `yaml:"urls"`
}

// LoadConfig loads the YAML config file.
// Note: the configuration is not checked for validity.
func LoadConfig(path string) (*config, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &config{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
