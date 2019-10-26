package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/gateway/config"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/middleware"
	"github.com/heetch/FabianG-technical-test/types"
	nsq "github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	responseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_response_time",
			Help:    "histogram of response times for gateway http handlers",
			Buckets: prometheus.ExponentialBuckets(0.5e-3, 2, 14), // 0.5ms to 4s
		},
		[]string{"path", "status_code"},
	)
)

func init() {
	prometheus.MustRegister(responseTimeHistogram)
}

func newGatewayHandler(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (http.Handler, error) {
	// initialize middleware common to all handlers
	var mw []middleware.Middleware
	mw = append(mw, middleware.NewRecoverHandler())
	mw = append(mw, middleware.NewContextLog(logger)...)
	// we measure response time for all handlers
	mc := middleware.NewMetricsConfig().WithTimeHist(responseTimeHistogram)
	mw = append(mw, middleware.NewMetricsHandler(mc))

	router := mux.NewRouter()
	for _, url := range cfg.URLs {
		h, err := newHandler(ctx, url, logger)
		if err != nil {
			return nil, err
		}
		// relies on valid URL configuration
		router.Handle(url.Path, middleware.Use(h, mw...)).Methods(url.Method)
	}
	router.Handle("/ready", &handler.ReadinessHandler{})
	return router, nil
}

func newHandler(ctx context.Context, u config.URL, logger zerolog.Logger) (http.Handler, error) {
	p, err := u.Protocol()
	if err != nil {
		return nil, err
	}
	switch p {
	case config.NSQ:
		return newNSQHandler(ctx, u, logger)
	case config.HTTP:
		// in a real world scenario we would factor this out to perform more
		// sophisticated operations like rewriting headers for HTTPS connections.
		// we ignore Transfer-Encoding hop-by-hop header; expecting `chunked` to
		// be applied if required. returns http.StatusBadGateway if backend is
		// not reachable.
		// TODO: add circuit-breaker
		return httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   u.HTTP.Host,
		}), nil
	default:
		return nil, fmt.Errorf("no handler found for %s", p)
	}
}

// nsqHandler transforms locations from http requests to nsq messages.
type nsqHandler struct {
	topic     string
	producers map[string]*nsq.Producer // safe for concurrent reads
}

func newNSQHandler(ctx context.Context, u config.URL, logger zerolog.Logger) (*nsqHandler, error) {
	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("go-nsq/%s", nsq.VERSION)

	// producers will lazily connect to the nsqd instance (and re-connect) when
	// Publish commands are executed. Note that throttling is not enabled
	producers := make(map[string]*nsq.Producer)
	for _, addr := range u.NSQ.TCPAddrs {
		producer, err := nsq.NewProducer(addr, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create nsq.Producer - %s", err)
		}

		// zerolog.Logger does not work here
		// producer.SetLogger(logger, logger.Level)

		producers[addr] = producer
	}

	// stop publishers on shutdown
	go func() {
		<-ctx.Done()
		for _, p := range producers {
			p.Stop()
		}
	}()

	return &nsqHandler{
		topic:     u.NSQ.Topic,
		producers: producers,
	}, nil
}

func (n *nsqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	// marshal instead of decode since we expect a single JSON string
	// only not a stream or additional data
	var l types.Location
	err = json.Unmarshal(body, &l)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	// relies on sane input for 'id'; currently sanitized by mux only
	l.ID = mux.Vars(r)["id"]
	b, err := json.Marshal(l)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	// publish synchronously
	// todo, requeue in case failure
	if err := hystrix.Do("publish_nsq", func() error { // circuit-breaker
		for _, producer := range n.producers {
			if err := producer.Publish(n.topic, b); err != nil {
				return err
			}
		}
		return nil
	}, nil); err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
