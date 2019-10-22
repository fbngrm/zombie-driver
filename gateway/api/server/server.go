package server

import (
	"context"
	"net/http"

	"github.com/heetch/FabianG-technical-test/gateway/api/config"
	"github.com/rs/zerolog"
)

type HTTPServer struct {
	server *http.Server
	logger zerolog.Logger
}

func New(ctx context.Context, addr string, cfg *config.Config, logger zerolog.Logger) (*HTTPServer, error) {
	router, err := newGatewayHandler(ctx, cfg, logger)
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
