package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/gateway/api"
	"github.com/heetch/FabianG-technical-test/gateway/api/config"
	"github.com/heetch/FabianG-technical-test/gateway/api/middleware"
	nsq "github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func newGatewayHandler(cfg *config.Config, logger zerolog.Logger) (http.Handler, error) {
	// initialize middleware common to all handlers
	var mw []middleware.Middleware
	mw = append(mw,
		middleware.NewAuthCheck(authtoken),
		middleware.NewRecoverHandler(),
	)
	mw = append(mw, middleware.NewContextLog(logger)...)
	// we measure response time only for all handlers
	mc := middleware.NewMetricsConfig().WithTimeHist(responseTimeHistogram)
	mw = append(mw, middleware.NewMetricsHandler(mc))

	router := mux.NewRouter()
	for _, url := range cfg.URLs {
		h, err := newHandler(url, logger)
		if err != nil {
			return nil, err
		}
		// NOTE: relies on valid URL configuration
		router.Handle(url.Path, middleware.Use(h, mw...)).Methods(url.Method)
	}
	router.Handle("/ready", &readinessHandler{})
	return router, nil
}

func newHandler(u config.URL, logger zerolog.Logger) (http.Handler, error) {
	p, err := u.Protocol()
	if err != nil {
		return nil, err
	}
	switch p {
	case config.NSQ:
		return newNSQHandler(u)
	case config.HTTP:
		// in a real world scenario we would factor this out to perform more
		// sofisticated operations like rewriting/removing headers e.g.
		// Authentication.
		// we ignore Transfer-Encoding hop-by-hop header; expecting `chunked` to
		// be applied if required. returns http.StatusBadGateway if backend is
		// not reachable.
		return httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   u.HTTP.Host,
		}), nil
	default:
		return nil, fmt.Errorf("no handler found for %s", p)
	}
}

type readinessHandler struct{}

func (h *readinessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(health())
}

type nsqHandler struct {
	topic     string
	producers map[string]*nsq.Producer
}

// NOTE: throttling is not enabled for producers which should be done in a
// production environment.
func newNSQHandler(u config.URL) (*nsqHandler, error) {
	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("go-nsq/%s", nsq.VERSION)
	producers := make(map[string]*nsq.Producer)
	for _, addr := range u.NSQ.TCPAddrs {
		producer, err := nsq.NewProducer(addr, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create nsq.Producer - %s", err)
		}
		producers[addr] = producer
	}
	return &nsqHandler{
		topic:     u.NSQ.Topic,
		producers: producers,
	}, nil
}

func (n *nsqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}
	var l api.Location
	// marshal instead of decode since we expect a single JSON string
	// only not a stream or additional data
	err = json.Unmarshal(body, &l)
	if err != nil {
		// should be expose error details here?
		writeError(w, r, err, http.StatusBadRequest)
		return
	}
	// relies on sane input for 'id'; currently sanitized by mux only
	l.ID = mux.Vars(r)["id"]
	b, err := json.Marshal(l)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}
	for _, producer := range n.producers {
		err := producer.Publish(n.topic, b)
		if err != nil {
			writeError(w, r, err, http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// Copied from https://github.com/heetch/regula/blob/master/api/server/handler.go
// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// encodeJSON encodes v to w in JSON format.
func encodeJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		loggerFromRequest(r).Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
	}
}

func loggerFromRequest(r *http.Request) *zerolog.Logger {
	logger := hlog.FromRequest(r).With().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Logger()
	return &logger
}

// writeError writes an error to the http response in JSON format.
func writeError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// Prepare log.
	logger := loggerFromRequest(r).With().
		Err(err).
		Int("status", code).
		Logger()

	// Hide error from client if it's internal.
	if code == http.StatusInternalServerError {
		logger.Error().Msg("unexpected http error")
		err = errInternal
	} else {
		logger.Debug().Msg("http error")
	}
	encodeJSON(w, r, &api.Error{Err: err.Error()}, code)
}
