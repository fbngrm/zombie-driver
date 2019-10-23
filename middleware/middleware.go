package middleware

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Middleware wraps handler functions
type Middleware func(http.Handler) http.Handler

// Use is the middleware chainer. Note, that the h will be the innermost
// handler, the first mws will be the innermost middleware handler, and the
// last mws will be the outermost middleware handler.
func Use(h http.Handler, mws ...Middleware) http.Handler {
	for _, m := range mws {
		if m != nil {
			h = m(h)
		}
	}
	return h
}

// NewAuthCheck produces middleware that authenticates the client.
func NewAuthCheck(token string) Middleware {
	authToken := "Bearer " + token
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(authToken)) != 1 {
				handler.WriteError(w, r, errors.New("Authentication is required"), http.StatusForbidden)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// NewRecoverHandler produces middleware for use as the last request handler
// in order to avoid the service completely crashing when there's a runtime
// panic.
func NewRecoverHandler() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					handler.WriteError(w, r, fmt.Errorf("PANIC: %+v", err), http.StatusInternalServerError)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}

// NewContextLog returns middleware that adds logger to request context.
func NewContextLog(logger zerolog.Logger) []Middleware {
	var mw []Middleware
	mw = append(mw, hlog.NewHandler(logger))
	// Install some provided extra handler to set some request's context fields.
	// Thanks to those handler, all our logs will come with some pre-populated fields.
	mw = append(mw, hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	mw = append(mw, hlog.RemoteAddrHandler("ip"))
	mw = append(mw, hlog.UserAgentHandler("user_agent"))
	mw = append(mw, hlog.RefererHandler("referer"))
	mw = append(mw, hlog.RequestIDHandler("req_id", "Request-Id"))
	return mw
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func newStatusResponseWriter(rw http.ResponseWriter) *statusResponseWriter {
	switch v := rw.(type) {
	case *statusResponseWriter:
		return v
	default:
		return &statusResponseWriter{ResponseWriter: rw}
	}
}

func (rw *statusResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// MetricsConfig keeps metrics configuration for use by MetricsHandler.
type MetricsConfig struct {
	reqGauge *prometheus.GaugeVec
	respCnt  *prometheus.CounterVec
	reqTime  *prometheus.HistogramVec
}

// NewMetricsConfig returns a new empty MetricsConfig structure.
func NewMetricsConfig() *MetricsConfig {
	return &MetricsConfig{}
}

// WithGauge adds gauge measuring how many requests are being handled at the
// moment. "path" label is set to the URL path.
func (mc *MetricsConfig) WithGauge(gauge *prometheus.GaugeVec) *MetricsConfig {
	mc.reqGauge = gauge
	return mc
}

// WithCounter adds counter that is updated after processing every request.
// "path" label is set to the URL path. "status_code" label is set on the
// counter to the status code of the response
func (mc *MetricsConfig) WithCounter(cnt *prometheus.CounterVec) *MetricsConfig {
	mc.respCnt = cnt
	return mc
}

// WithTimeHist adds histogram that measures distribution of the response
// times. "path" label is set to the URL path. "status_code" label is set on
// the counter to the status code of the response
func (mc *MetricsConfig) WithTimeHist(hist *prometheus.HistogramVec) *MetricsConfig {
	mc.reqTime = hist
	return mc
}

// NewMetricsHandler returns middleware that updates prometheus metrics.
func NewMetricsHandler(mc *MetricsConfig) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			sw := newStatusResponseWriter(w)
			l := prometheus.Labels{"path": r.URL.Path}
			if mc.reqGauge != nil {
				mc.reqGauge.With(l).Inc()
			}
			h.ServeHTTP(sw, r)
			if mc.reqGauge != nil {
				mc.reqGauge.With(l).Dec()
			}
			l["status_code"] = fmt.Sprint(sw.status)
			if mc.respCnt != nil {
				mc.respCnt.With(l).Inc()
			}
			if mc.reqTime != nil {
				mc.reqTime.With(l).Observe(time.Since(startTime).Seconds())
			}
		})
	}
}
