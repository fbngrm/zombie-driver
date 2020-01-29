package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

var testURLs = []string{
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
    method: "GET"`, // second case; one url; missing protocol
}

// map expected errors to expected output
type res struct {
	u URL
	p protocol
	e error // expected error
}

var cfgtests = []struct {
	d    string // test case description
	in   string // input
	want []res  // expected result
}{
	{
		d:  "expect config to contain 2 URLs",
		in: testURLs[0],
		want: []res{
			res{
				u: URL{
					Path:   "/drivers/{id:[0-9]+}/locations",
					Method: "PATCH",
					NSQ: NSQConf{
						Topic:    "locations",
						TCPAddrs: []string{"127.0.0.1:4150"},
					},
				},
				p: NSQ,
				e: nil,
			},
			res{
				u: URL{
					Path:   "/drivers/{id:[0-9]+}",
					Method: "GET",
					HTTP: HTTPConf{
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
		in: testURLs[1],
		want: []res{
			res{
				u: URL{
					Path:   "/drivers/{id:[0-9]+}",
					Method: "GET",
				},
				e: errors.New("URL is missing protocol"),
			},
		},
	},
}

func TestLoad(t *testing.T) {
	for i, tt := range cfgtests {
		r := strings.NewReader(testURLs[i])
		cfg, err := load(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w, g := len(tt.want), len(cfg.URLs); w != g {
			t.Fatalf("%s: want count: %d got: %d", tt.d, w, g)
		}
		for i, u := range cfg.URLs {
			if w, g := tt.want[i].u, u; !reflect.DeepEqual(w, g) {
				t.Errorf("%s:\nwant URL:\n%+v\ngot:\n%+v", tt.d, w, g)
			}

			p, err := u.Protocol()
			// unexpected error
			if w, g := tt.want[i].e, err; w == nil && g != nil {
				t.Fatalf("%s: unexpected error: %s", tt.d, g.Error())
			}
			// expected error
			if w, g := tt.want[i].e, err; w != nil && g == nil {
				t.Fatalf("%s: want error: %v got: %v", tt.d, w, g)
			}
			if w, g := tt.want[i].e, err; (w != nil && g != nil) && (w.Error() != g.Error()) {
				t.Fatalf("%s: want error: %v got: %v", tt.d, w.Error(), g.Error())
			}

			if w, g := tt.want[i].p, p; w != g {
				t.Errorf("%s: want protocol: %s got:%s", tt.d, w, g)
			}
		}
	}
}
