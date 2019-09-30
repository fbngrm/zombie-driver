package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type HTTPServer struct {
	server *http.Server
	logger zerolog.Logger
}

func New(port int, cfg *config, logger zerolog.Logger) (*HTTPServer, error) {
	router := mux.NewRouter()
	for _, url := range cfg.URLs {
		h, err := newHandler(url, logger)
		if err != nil {
			return nil, err
		}
		// NOTE: relies on valid configuration
		router.Handle(url.Path, h).Methods(url.Method)
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

func (s *HTTPServer) Run(ctx context.Context) error {
	go func() {
		s.logger.Info().Msgf("Listening on %s", s.server.Addr)
		err := s.server.ListenAndServe()
		if err != http.ErrServerClosed {
			s.logger.Fatal().Msgf("http server exited with error: %v", err)
		} else {
			s.logger.Info().Msgf("http server has closed")
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to shutdown server")
	}
	return err
}
