package cli

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/heetch/FabianG-technical-test/gateway/api/server"
	"github.com/rs/zerolog"
)

// Using default log level debug and write to stderr.
// Note: We log in (inefficient) human friendly format to console here since it
// is a coding challenge. In a production environment we would prefer structured,
// machine parsable format. So we could make use of automated log analysis e.g.
// error reporting.
func NewLogger() zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	logger = logger.Output(zerolog.ConsoleWriter{Out: f})
	log.SetFlags(0)
	log.SetOutput(logger)
	return logger
}

func RunServer(srv *server.Server, addr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		<-quit
		cancel()
	}()

	return srv.Run(ctx, addr)
}
