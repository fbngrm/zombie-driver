package server

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	responseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_response_time",
			Help:    "histogram of response times for zombie driver http handlers",
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

func New(addr, driverLocationURL string, zombieRadius float64, zombieTime int, logger zerolog.Logger) (*HTTPServer, error) {
	router, err := newZombieHandler(driverLocationURL, zombieRadius, zombieTime, logger)
	if err != nil {
		return nil, err
	}
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	return &HTTPServer{
		server: server,
		logger: logger,
	}, nil
}

func (s *HTTPServer) Run() {
	s.logger.Info().Msgf("http server listening on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Fatal().Err(err).Msg("http server exited with error")
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) {
	s.logger.Info().Msg("shutting down http server down")

	// this stops accepting new requests and waits for the running ones to
	// finish before returning. See net/http docs for details.
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("http server shutdown error")
	}
}
