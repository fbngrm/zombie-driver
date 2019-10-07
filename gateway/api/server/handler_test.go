package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
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

var proxytest = handlertest{
	d: "test HTTP reverse proxy",
	c: config.Config{
		URLs: []config.URL{
			config.URL{
				Path:   "/drivers/{id:[0-9]+}/locations",
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
			p: "/drivers/1/locations",
			m: "GET",
			s: http.StatusMethodNotAllowed,
		},
		proxyreq{
			d: "expect missing auth header",
			p: "/drivers/1/locations",
			m: "PATCH",
			s: http.StatusForbidden,
		},
		proxyreq{
			d: "expect status success",
			p: "/drivers/1/locations",
			m: "PATCH",
			s: http.StatusOK,
			// NOTE: in a real world scenario we would never hardcode an auth
			// token, not even a test token.
			// This should be loaded from e.g. an env file or a secret. We would
			// not check-in env files into a remote git repo.
			h: map[string]string{"Authorization": "Bearer AUTH_TOKEN"},
		},
	},
}

func TestProxy(t *testing.T) {
	// create a target for the reverse proxy
	backendResponse := "I am the backend"
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(backendResponse))
	}))
	defer backend.Close()

	// NOTE: we need to overwrite the URL Host in the test config here, which
	// is not a desired behaviour normally. The httptest.Server uses a local
	// Listener initialized to listen on a random port. Using a custom Listener
	// and providing a port would require supporting `serveFlag` and IPv6.
	// For more info see:
	// https://golang.org/src/net/http/httptest/server.go?s=477:1449#L72
	u, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", proxytest.d, err)
	}
	proxytest.c.URLs[0].HTTP.Host = u.Host
	// create the proxy handler to test
	h, err := newGatewayHandler(&proxytest.c, zerolog.New(ioutil.Discard))
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", proxytest.d, err)
	}

	frontend := httptest.NewServer(h)
	defer frontend.Close()
	frontendClient := frontend.Client()

	for _, tt := range proxytest.cases {
		t.Run(tt.d, func(t *testing.T) {
			req, _ := http.NewRequest(tt.m, frontend.URL+tt.p, nil)
			req.Close = true
			req.Host = "test-host"
			for k, v := range tt.h {
				req.Header.Set(k, v)
			}
			req.Header.Set("Connection", "close")
			res, err := frontendClient.Do(req)
			if err != nil {
				t.Fatalf("%s: unexpected error %v", tt.d, err)
			}
			if res.StatusCode != tt.s {
				t.Errorf("%s: expect status code %d got %d", tt.d, tt.s, res.StatusCode)
			}
		})
	}
}
