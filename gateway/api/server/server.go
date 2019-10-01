package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/gateway/api/middleware"
	"github.com/rs/zerolog"
)

// FIXME: in a real world scenario we would never hardcode a token here!
// Instead it should be loaded from an encrypted env configuration. For the
// sake of simplicity in a coding challenge, we violate this rule.
var authtoken = "AUTH_TOKEN"

type HTTPServer struct {
	server *http.Server
	logger zerolog.Logger
}

func New(port int, cfg *config, logger zerolog.Logger) (*HTTPServer, error) {
	router, err := newGatewayHandler(cfg, logger)
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
	s.logger.Info().Msgf("listening on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Fatal().Err(err).Msg("http server exited with error")
	}
}

func (s *HTTPServer) Shutdown() error {
	s.logger.Info().Msg("shutting down HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// this stops accepting new requests and waits for the running ones to
	// finish before returning. See net/http docs for details.
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("http server shutdown error")
		return err
	}
	return nil
}

func newGatewayHandler(cfg *config, logger zerolog.Logger) (http.Handler, error) {
	// initialize middleware common to all handlers
	var mw []middleware.Middleware
	mw = append(mw,
		middleware.NewAuthCheck(authtoken),
		middleware.NewRecoverHandler(),
	)
	mw = append(mw, middleware.NewContextLog(logger)...)

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
