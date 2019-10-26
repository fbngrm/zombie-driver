package server

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/heetch/FabianG-technical-test/gateway/config"
	nsq "github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
)

// urls for test handlers
var gatewayConf = config.Config{
	URLs: []config.URL{
		config.URL{ // nsq
			Path:   "/drivers/{id:[0-9]+}/locations",
			Method: "PATCH",
			NSQ: config.NSQConf{
				Topic:    "test-locations",
				TCPAddrs: []string{"127.0.0.1:4150"},
			},
		},
		config.URL{ // proxy
			Path:   "/drivers/{id:[0-9]+}",
			Method: "GET",
			HTTP: config.HTTPConf{
				Host: "zombie-driver", // will be overwritten by test server host and port
			},
		},
	},
}

var gatewayTests = map[string]struct {
	d string // description of test case
	z string // response of zombie-driver srevice mock
	p string // request path
	r string // expected response data
	s int    // expected response status code
}{
	"0": {
		d: "expect StatusBadGateway when failing to reach the backend",
		p: "/drivers/0", // 0 test hijacked requests
		s: http.StatusBadGateway,
	},
	"1": {
		d: "expect successful proxying; #1",
		z: `{"id":1,"zombie":true}`,
		p: "/drivers/1",
		r: `{"id":1,"zombie":true}`,
		s: http.StatusOK,
	},
	"2": {
		d: "expect successful proxying; #2",
		z: `{"id":2,"zombie":false}`,
		p: "/drivers/2",
		r: `{"id":2,"zombie":false}`,
		s: http.StatusOK,
	},
	"3": {
		d: "expect StatusNotFound for invalid URL",
		p: "/drivers",
		r: "404 page not found",
		s: http.StatusNotFound,
	},
	"4": {
		d: "expect StatusNotFound for unknown driver",
		p: "/drivers/404",
		s: http.StatusNotFound,
	},
}

func TestProxy(t *testing.T) {
	// mute logger in tests
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)

	// create a mock zombie-service for the reverse proxy
	zombieService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// driver id
		segments := strings.Split(r.URL.Path, "/")
		if len(segments) != 3 {
			t.Fatalf("expect 3 path segments but got %d", len(segments))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		id := segments[2]

		// not reachable
		if id == "0" {
			c, _, _ := w.(http.Hijacker).Hijack()
			c.Close()
			return
		}

		// we ignore other hop-by-hop headers for now
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Error("expect X-Forwarded-For header")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// send mock data
		if p, ok := gatewayTests[id]; ok {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(p.z))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
				return
			}
			return
		}
		// driver ID unknown
		w.WriteHeader(http.StatusNotFound)
	}))
	defer zombieService.Close()

	// Note, we need to overwrite the URL Host in the test config with the
	// address of the test zombieService. The httptest.Server uses a local
	// Listener initialized to listen on a random port. Using a custom Listener
	// and providing a port would require supporting `serveFlag` and IPv6.
	// For more info see:
	// https://golang.org/src/net/http/httptest/server.go?s=477:1449#L72
	u, err := url.Parse(zombieService.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gatewayConf.URLs[1].HTTP.Host = u.Host

	// handler to test
	h, err := newGatewayHandler(context.Background(), &gatewayConf, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// serve test handler
	gatewayService := httptest.NewServer(h)
	defer gatewayService.Close()
	gatewayClient := gatewayService.Client()

	t.Run("zombie-service", func(t *testing.T) {
		for id := range gatewayTests {
			tt := gatewayTests[id]
			t.Run(tt.d, func(t *testing.T) {
				t.Parallel()
				req, err := http.NewRequest("GET", gatewayService.URL+tt.p, nil)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				req.Close = true
				req.Header.Set("Connection", "close")

				res, err := gatewayClient.Do(req)
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}

				if w, g := tt.s, res.StatusCode; w != g {
					t.Errorf("want status code %d got %d", w, g)
				}
				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("failed to read response %v", err)
				}
				if w, g := tt.r, strings.TrimSpace(string(data)); w != g {
					t.Errorf("want response %s got %s", w, g)
				}
			})
		}
	})
}

var nsqTests = []struct {
	d string // description of test case
	b string // payload of test test request
	p string // request path
	s int    // expected response status code
}{
	{
		d: "expect StatusBadRequest for malformatted JSON payload",
		b: `"id":"0","latitude":48.864193,"longitude":2.350498}`,
		p: "/drivers/0/locations",
		s: http.StatusBadRequest,
	},
	{
		d: "expect StatusBadRequest for empty payload",
		p: "/drivers/0/locations",
		s: http.StatusBadRequest,
	},
	{
		d: "expect StatusNotFound for invalid user",
		p: "/drivers/foo/locations",
		s: http.StatusNotFound,
	},
	{
		d: "expect StatusOK; #1",
		b: `{"id":"0","latitude":48.864193,"longitude":2.350498}`,
		p: "/drivers/0/locations",
		s: http.StatusOK,
	},
	{
		d: "expect StatusOK; #2",
		b: `{"id":"1","latitude":48.864193,"longitude":2.350498}`,
		p: "/drivers/1/locations",
		s: http.StatusOK,
	},
	{
		d: "expect StatusOK; #3",
		b: `{"id":"2","latitude":48.864193,"longitude":2.350498}`,
		p: "/drivers/2/locations",
		s: http.StatusOK,
	},
	{
		d: "expect StatusOK; #4",
		b: `{"id":"3","latitude":48.864193,"longitude":2.350498}`,
		p: "/drivers/3/locations",
		s: http.StatusOK,
	},
	{
		d: "expect StatusOK; #5",
		b: `{"id":"123456789","latitude":48.864193,"longitude":2.350498}`,
		p: "/drivers/123456789/locations",
		s: http.StatusOK,
	},
	{
		d: "expect StatusOK; #6",
		b: `{"id":"10000000","latitude":48.864193,"longitude":2.350498}`,
		p: "/drivers/10000000/locations",
		s: http.StatusOK,
	},
}

var nullLogger = log.New(ioutil.Discard, "", log.LstdFlags)

// test needs to set a timeout
func TestNSQ(t *testing.T) {
	// mute logger in tests
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)

	// handler to test
	h, err := newGatewayHandler(context.Background(), &gatewayConf, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gatewayService := httptest.NewServer(h)
	defer gatewayService.Close()
	gatewayClient := gatewayService.Client()

	count := 0                // expected message count
	want := make([]string, 0) // expected messages

	// send requests to publish nsq messages
	t.Run("nsq-server", func(t *testing.T) {
		for id := range nsqTests {
			tt := nsqTests[id]
			t.Run(tt.d, func(t *testing.T) {
				req, err := http.NewRequest("PATCH", gatewayService.URL+tt.p, nil)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				req.Close = true
				req.Header.Set("Connection", "close")
				req.Body = ioutil.NopCloser(strings.NewReader(tt.b))
				res, err := gatewayClient.Do(req)
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				if w, g := tt.s, res.StatusCode; w != g {
					t.Errorf("want status code %d got %d", w, g)
				}
				// increase counter only if NSQ message has been sent successfully
				if res.StatusCode == http.StatusOK {
					want = append(want, tt.b)
					count++
				}
			})
		}
	})

	// no messages sent
	if count < 1 {
		return
	}

	// need to set test timeout; potentially blocks forever if not
	// all messages are delivered
	c := newConsumerHandler(t, gatewayConf.URLs[0].NSQ.Topic, count)
	err = c.read(gatewayConf.URLs[0].NSQ.TCPAddrs[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w, g := len(want), len(c.got); w != g {
		t.Errorf("want %d messages got %d", w, g)
	}
	for i := range want {
		if w, g := want[i], c.got[i]; w != g {
			t.Errorf("want message %s got %s", w, g)
		}
	}
}

type ConsumerHandler struct {
	t    *testing.T
	want int
	got  []string
	c    *nsq.Consumer
}

func newConsumerHandler(t *testing.T, topic string, want int) *ConsumerHandler {
	config := nsq.NewConfig()
	config.DefaultRequeueDelay = 0
	config.MaxBackoffDuration = 50 * time.Millisecond
	c, _ := nsq.NewConsumer(topic, "ch", config)
	c.SetLogger(nullLogger, nsq.LogLevelInfo)

	h := &ConsumerHandler{
		t:    t,
		want: want,
		got:  make([]string, 0),
		c:    c,
	}
	c.AddHandler(h)
	return h
}

func (h *ConsumerHandler) HandleMessage(message *nsq.Message) error {
	h.got = append(h.got, string(message.Body))
	if len(h.got) == h.want {
		h.c.Stop()
	}
	return nil
}

func (h *ConsumerHandler) read(addr string) error {
	if h.want == 0 {
		return errors.New("want < 1")
	}
	err := h.c.ConnectToNSQD(addr)
	if err != nil {
		return err
	}
	<-h.c.StopChan
	return nil
}
