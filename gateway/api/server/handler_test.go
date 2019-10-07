package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
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
	i string            // driver id
	h map[string]string // header of test request
	s int               // expected response status code
}

type handlertest struct {
	d     string            // description of test
	p     map[string]string // test request payloads by driver ID
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
var proxytest = handlertest{
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
	p: map[string]string{
		"0": "",
		"1": `{"latitude": 48.864193,"longitude": 2.350498}`,
		"2": "not reachable",
	},
	cases: []proxyreq{
		proxyreq{
			d: "should accept PATCH only",
			p: "/drivers",
			i: "0", // empty payload
			m: "GET",
			s: http.StatusMethodNotAllowed,
		},
		proxyreq{
			d: "expect missing auth header",
			p: "/drivers",
			i: "0", // empty payload
			m: "PATCH",
			s: http.StatusForbidden,
		},
		proxyreq{
			d: "expect status success",
			p: "/drivers",
			i: "1", // test payload
			m: "PATCH",
			s: http.StatusOK,
			h: map[string]string{"Authorization": "Bearer AUTH_TOKEN"}, // don't do this
		},
		proxyreq{
			d: "expect StatusBadGateway when failing to reach the backend",
			p: "/drivers",
			i: "2", // -1 tests hijacked requests
			m: "PATCH",
			s: http.StatusBadGateway,
			h: map[string]string{"Authorization": "Bearer AUTH_TOKEN"}, // don't do this
		},
	},
}

func TestProxy(t *testing.T) {
	// mute logger in tests
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)

	// create a backend for the reverse proxy
	backendResponse := "I am the backend"

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we use URL path to test hijacked requests/unreachable backend
		segments := strings.Split(r.URL.Path, "/")
		if len(segments) != 3 {
			t.Fatalf("expect 3 path segments but got %d", len(segments))
		}
		id := segments[2]
		if id == "2" { // backend not reachable
			c, _, _ := w.(http.Hijacker).Hijack()
			c.Close()
			return
		}
		// we ignore Transfer-Encoding hop-by-hop header; expecting chunked to
		// be applied if required
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Errorf("%s: didn't get X-Forwarded-For header", proxytest.d)
			w.WriteHeader(http.StatusBadRequest)
		}
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		payload := proxytest.p[id]
		if g := string(bodyBytes); g != payload {
			t.Errorf("%s: got body %q; expected %q", proxytest.d, g, payload)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte(backendResponse))
	}))
	defer backend.Close()

	// NOTE: we need to overwrite the URL Host in the test config with the
	// address of the test backend here. The httptest.Server uses a local
	// Listener initialized to listen on a random port. Using a custom Listener
	// and providing a port would require supporting `serveFlag` and IPv6.
	// For more info see:
	// https://golang.org/src/net/http/httptest/server.go?s=477:1449#L72
	u, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", proxytest.d, err)
	}
	// overwrite test config
	proxytest.c.URLs[0].HTTP.Host = u.Host
	// create the proxy handler to test
	h, err := newGatewayHandler(&proxytest.c, logger)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", proxytest.d, err)
	}

	frontend := httptest.NewServer(h)
	defer frontend.Close()
	frontendClient := frontend.Client()

	for _, tc := range proxytest.cases {
		t.Run(tc.d, func(t *testing.T) {
			req, _ := http.NewRequest(tc.m, frontend.URL+path.Join(tc.p, tc.i), nil)
			req.Close = true // close TCP conn after response was read
			req.Host = "test-host"
			for k, v := range tc.h {
				req.Header.Set(k, v)
			}
			req.Header.Set("Connection", "close")
			req.Body = ioutil.NopCloser(strings.NewReader(proxytest.p[tc.i]))
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
