package server

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
)

// HTTPServer represents an http server.
type HTTPServer struct {
	server *http.Server
	logger zerolog.Logger
}

// New returns an HTTPServer instance with a locationHandler.
func New(httpAddr string, rf RangeFetcher, logger zerolog.Logger) (*HTTPServer, error) {
	router, err := newLocationHandler(rf, logger)
	if err != nil {
		return nil, err
	}
	server := &http.Server{
		Addr:    httpAddr,
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
	s.logger.Info().Msg("shutting down http server")

	// this stops accepting new requests and waits for the running ones to
	// finish before returning. See net/http docs for details.
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("http server shutdown error")
	}
}
