package metrics

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// MetricsServer provides an endpoint for a prometheus scraper.
type MetricsServer struct {
	cnt    uint64
	srv    *http.Server
	logger zerolog.Logger
}

func New(addr string, logger zerolog.Logger) *MetricsServer {
	ms := &MetricsServer{
		logger: logger,
	}
	promHandler := promhttp.Handler()
	cntHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			promHandler.ServeHTTP(w, req)
			atomic.AddUint64(&ms.cnt, 1)
		})
	mux := http.NewServeMux()
	mux.Handle("/metrics", cntHandler)
	ms.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return ms
}

func (ms *MetricsServer) Run() {
	ms.logger.Info().Msgf("metrics server listening on %s", ms.srv.Addr)
	if err := ms.srv.ListenAndServe(); err != http.ErrServerClosed {
		ms.logger.Fatal().Err(err).Msg("metrics server exited with error")
	}
}

func (ms *MetricsServer) Shutdown(ctx context.Context) {
	ms.logger.Info().Msg("waiting for metrics to be scraped")

	// we're getting the current count of accesses to /metrics and waiting for
	// it to increase, i.e. we're waiting till prometheus will scrape the final
	// metrics before exiting
	n := atomic.LoadUint64(&ms.cnt)
	t := time.NewTicker(100 * time.Millisecond)
LOOP:
	for {
		select {
		case <-t.C:
			cnt := atomic.LoadUint64(&ms.cnt)
			if cnt > n {
				break LOOP
			}
		case <-ctx.Done():
			break LOOP
		}
	}
	t.Stop()
	ms.logger.Info().Msg("shutting metrics server down")
	if err := ms.srv.Shutdown(ctx); err != nil {
		ms.logger.Error().Err(err).Msg("metrics server shutdown error")
	}
}
