package cli

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/heetch/FabianG-technical-test/gateway/server"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/metrics"
	"github.com/rs/zerolog"
)

// Using default log level debug and write to stderr.
// Note, we log in (inefficient) human friendly format to console here since it
// is a coding challenge. In a production environment we would prefer structured,
// machine parsable format. So we could make use of automated log analysis e.g.
// error reporting.
func NewLogger(service, version string) zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// replace standard log
	log.SetFlags(0)
	log.SetOutput(logger)
	return logger.With().
		Interface("service", service).
		Interface("version", version).
		Logger()
}

func RunServer(ctx context.Context, httpSrv *server.HTTPServer, metricsSrv *metrics.MetricsServer, shutdownDelay int) {
	go httpSrv.Run()
	go metricsSrv.Run()

	<-ctx.Done()

	handler.HealthCheckShutDown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownDelay)*time.Millisecond)
	defer cancel()

	// when shutting down, we first gracefully shutting down the main http
	// server, waiting for it to finish processing all the running requests,
	// then we shut down the metrics server, which includes waiting for
	// prometheus to scrape the metrics one more time, to avoid loosing any data.
	httpSrv.Shutdown(ctx)
	metricsSrv.Shutdown(ctx)
}
