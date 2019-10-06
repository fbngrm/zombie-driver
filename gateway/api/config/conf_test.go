package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

var inputs = []string{
	`urls:
  -
    path: "/drivers/{id:[0-9]+}/locations"
    method: "PATCH"
    nsq:
      topic: "locations"
      dest_tcp_addr:
          - "127.0.0.1:4150"
  -
    path: "/drivers/{id:[0-9]+}"
    method: "GET"
    http:
      host: "zombie-driver"`, // first case; two urls
	`urls:
  -
    path: "/drivers/{id:[0-9]+}"
    method: "GET"`, // second case; one url
}

type testURL struct {
	Path   string
	Method string
	NSQ    struct {
		Topic    string
		TCPAddrs []string
	}
	HTTP struct {
		Host string
	}
}

// map expected errors to expected output
type res struct {
	u testURL
	p protocol
	e error // expected error
}

var cfgtests = []struct {
	d    string // description
	in   string // input
	want []res  // expected result
}{
	{
		d:  "expect config to contain two URLs",
		in: inputs[0],
		want: []res{
			res{
				u: testURL{
					Path:   "/drivers/{id:[0-9]+}/locations",
					Method: "PATCH",
					NSQ: struct {
						Topic    string
						TCPAddrs []string
					}{
						Topic:    "locations",
						TCPAddrs: []string{"127.0.0.1:4150"},
					},
					HTTP: struct {
						Host string
					}{},
				},
				p: NSQ,
				e: nil,
			},
			res{
				u: testURL{
					Path:   "/drivers/{id:[0-9]+}",
					Method: "GET",
					NSQ: struct {
						Topic    string
						TCPAddrs []string
					}{
						Topic:    "locations",
						TCPAddrs: []string{"127.0.0.1:4150"},
					},
					HTTP: struct {
						Host string
					}{
						Host: "zombie-driver",
					},
				},
				p: HTTP,
				e: nil,
			},
		},
	},
	{
		d:  "expect missing protocol error",
		in: inputs[1],
		want: []res{
			res{
				u: testURL{
					Path:   "/drivers/{id:[0-9]+}/locations",
					Method: "PATCH",
				},
				e: errors.New("URL is missing protocol"),
			},
		},
	},
}

func TestLoad(t *testing.T) {
	for i, tt := range cfgtests {
		r := strings.NewReader(inputs[i])
		cfg, err := load(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tt.want) != len(cfg.URLs) {
			t.Fatalf("%s: expected URL count: %d got: %d", tt.d, len(tt.want), len(cfg.URLs))
		}
		for i, u := range cfg.URLs {
			if reflect.DeepEqual(tt.want[i].u, u) {
				t.Errorf("%s: expected URL:\n%+v\ngot:\n%+v", tt.d, tt.want[i].u, u)
			}
			p, err := u.Protocol()
			if tt.want[i].e != nil && tt.want[i].e.Error() != err.Error() {
				t.Errorf("%s: expected error:\n%s\ngot:\n%s", tt.d, tt.want[i].e.Error(), err.Error())
			}
			if tt.want[i].p != p {
				t.Errorf("expected protocol: %s got:%s", tt.want[i].p, p)
			}
		}
	}
}
