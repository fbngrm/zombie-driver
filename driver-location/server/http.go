package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	// FIXME: in a real world scenario we would never hardcode a token here!
	// Instead it should be loaded from an encrypted env configuration. For the
	// sake of simplicity in a coding challenge, we violate this rule.
	authtoken             = "AUTH_TOKEN"
	responseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "driver_location_response_time",
			Help:    "histogram of response times for driver-location http handler",
			Buckets: prometheus.ExponentialBuckets(0.5e-3, 2, 14), // 0.5ms to 4s
		},
		[]string{"path", "status_code"},
	)
)

func init() {
	prometheus.MustRegister(responseTimeHistogram)
}

type HTTPServer struct {
	server *http.Server
	logger zerolog.Logger
}

func New(port int, logger zerolog.Logger) (*HTTPServer, error) {
	router, err := newLocationHandler(logger)
	if err != nil {
		return nil, err
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	return &HTTPServer{
		server: server,
		logger: logger,
	}, nil
}

func (s *HTTPServer) Run() {
	s.logger.Info().Msgf("driver-location: listening on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Fatal().Err(err).Msg("driver-location: http server exited with error")
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) {
	s.logger.Info().Msg("driver-location: shutting down HTTP server down")

	// this stops accepting new requests and waits for the running ones to
	// finish before returning. See net/http docs for details.
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("driver-location: http server shutdown error")
	}
}
