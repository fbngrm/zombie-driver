package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/heetch/FabianG-technical-test/gateway/api/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type proxyreq struct {
	d string            // description of test case
	p string            // URL path of test requests
	m string            // HTTP method of test requests
	h map[string]string // header of test request
	s int               // expected status code
}

type handlertest struct {
	d     string // description of test
	c     config.Config
	cases []proxyreq
}

var healthtest = handlertest{
	d: "test readiness handler",
	c: config.Config{},
	cases: []proxyreq{
		proxyreq{
			d: "expect status code 200",
			p: "/ready",
			m: "GET",
			s: 200,
		},
	},
}

func TestHealth(t *testing.T) {
	log := zerolog.New(ioutil.Discard)
	h, err := newGatewayHandler(&healthtest.c, log)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", healthtest.d, err)
	}
	for _, tt := range healthtest.cases {
		t.Run(tt.d, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.m, tt.p, nil)
			h.ServeHTTP(w, r)
			require.Equal(t, tt.s, w.Code, tt.d)
		})
	}
}

// Test data for reverse proxy tests. Requires to work with authentication
// middleware.
// FIXME: in a real world scenario we would never hardcode an auth token, not
// even a test token! This should be loaded from e.g. an env file or a secret.
// We would not check-in env files into a remote git repo. It is implemented
// here due to time limitations during coding challenge/prototyping.
var proxytests = []handlertest{
	handlertest{
		d: "test HTTP reverse proxy",
		c: config.Config{
			URLs: []config.URL{
				config.URL{
					Path:   "/drivers/{id:[0-9]+}",
					Method: "PATCH",
					HTTP: struct {
						Host string `yaml:"host"`
					}{
						Host: "", // will be overwritten by test server host and port
					},
				},
			},
		},
		cases: []proxyreq{
			proxyreq{
				d: "should accept PATCH only",
				p: "/drivers/1",
				m: "GET",
				s: http.StatusMethodNotAllowed,
			},
			proxyreq{
				d: "expect missing auth header",
				p: "/drivers/1",
				m: "PATCH",
				s: http.StatusForbidden,
			},
			proxyreq{
				d: "expect status success",
				p: "/drivers/1",
				m: "PATCH",
				s: http.StatusOK,
				h: map[string]string{"Authorization": "Bearer AUTH_TOKEN"}, // don't do this
			},
		},
	},
	handlertest{
		d: "test unreachable backend",
		c: config.Config{
			URLs: []config.URL{
				config.URL{
					Path:   "/drivers/{id:[0-9]+}/{mode:[a-z]*}",
					Method: "PATCH",
					HTTP: struct {
						Host string `yaml:"host"`
					}{
						Host: "", // will be overwritten by test server host and port
					},
				},
			},
		},
		cases: []proxyreq{
			proxyreq{
				d: "expect StatusBadGateway when failing to reach the backend",
				p: "/drivers/1/hangup",
				m: "PATCH",
				s: http.StatusBadGateway,
				h: map[string]string{"Authorization": "Bearer AUTH_TOKEN"}, // don't do this
			},
		},
	},
}

func TestProxy(t *testing.T) {
	// create a target for the reverse proxy
	backendResponse := "I am the backend"
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we use URL path to test hijacked requests/unreachable backend
		segments := strings.Split(r.URL.Path, "/")
		if len(segments) > 0 && segments[len(segments)-1] == "hangup" {
			c, _, _ := w.(http.Hijacker).Hijack()
			c.Close()
			return
		}
		w.Write([]byte(backendResponse))
	}))
	defer backend.Close()

	// mute logger in tests
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)

	for _, tt := range proxytests {
		// NOTE: we need to overwrite the URL Host in the test config here, with
		// the address of the test backend. The httptest.Server uses a local
		// Listener initialized to listen on a random port. Using a custom Listener
		// and providing a port would require supporting `serveFlag` and IPv6.
		// For more info see:
		// https://golang.org/src/net/http/httptest/server.go?s=477:1449#L72
		u, err := url.Parse(backend.URL)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.d, err)
		}
		// overwrite test config
		tt.c.URLs[0].HTTP.Host = u.Host
		// create the proxy handler to test
		h, err := newGatewayHandler(&tt.c, logger)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.d, err)
		}

		frontend := httptest.NewServer(h)
		defer frontend.Close()
		frontendClient := frontend.Client()

		for _, tc := range tt.cases {
			t.Run(tc.d, func(t *testing.T) {
				req, _ := http.NewRequest(tc.m, frontend.URL+tc.p, nil)
				req.Close = true // close TCP conn after response was read
				req.Host = "test-host"
				for k, v := range tc.h {
					req.Header.Set(k, v)
				}
				req.Header.Set("Connection", "close")
				res, err := frontendClient.Do(req)
				if err != nil {
					t.Fatalf("%s: unexpected error %v", tc.d, err)
				}
				if res.StatusCode != tc.s {
					t.Errorf("%s: expect status code %d got %d", tc.d, tc.s, res.StatusCode)
				}
			})
		}
	}
}
