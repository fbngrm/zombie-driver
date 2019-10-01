package server

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type MetricsServer struct {
	srv *http.Server
	cnt uint64
	log zerolog.Logger
}

func NewMetrics(port int, log zerolog.Logger) *MetricsServer {
	ms := &MetricsServer{
		log: log,
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
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return ms
}

func (ms *MetricsServer) Run() {
	ms.log.Info().Msgf("listening on %s", ms.srv.Addr)
	if err := ms.srv.ListenAndServe(); err != http.ErrServerClosed {
		ms.log.Fatal().Err(err).Msg("metrics server exited with error")
	}
}

func (ms *MetricsServer) Shutdown(ctx context.Context) {
	ms.log.Info().Msg("waiting for metrics to be scraped")

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
	ms.log.Info().Msg("shutting metrics server down")
	if err := ms.srv.Shutdown(ctx); err != nil {
		ms.log.Error().Err(err).Msg("metrics server shutdown error")
	}
}
