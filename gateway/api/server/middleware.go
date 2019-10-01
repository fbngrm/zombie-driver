package server

import (
	"crypto/subtle"
	"net/http"
	"time"

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

// NewAuthCheck produces middleware that authenticates the client
func NewAuthCheck(token string) Middleware {
	authToken := "Bearer " + token
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(authToken)) != 1 {
				http.Error(w, "Authentication is required", http.StatusForbidden)
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
					zerolog.Ctx(r.Context()).Error().Msgf("PANIC: %+v", err)
					http.Error(w, "Internal Server Error", 500)
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
