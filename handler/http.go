package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Copied from https://github.com/heetch/regula/blob/master/api/server/handler.go

// HTTP errors
var (
	errInternal   = errors.New("internal_error")
	errBadRequest = errors.New("bad_request")
)

// WriteError writes an error to the http response in JSON format.
func WriteError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// Prepare log.
	logger := loggerFromRequest(r).With().
		Err(err).
		Int("status", code).
		Logger()

	// Hide error from client if it's internal.
	if code == http.StatusInternalServerError {
		logger.Error().Msg("unexpected http error")
		err = errInternal
	} else if code == http.StatusBadRequest { // added
		logger.Error().Msg("http error bad request")
		err = errBadRequest
	} else {
		logger.Debug().Msg("http error")
	}
	EncodeJSON(w, r, &Error{Err: err.Error()}, code)
}

// EncodeJSON encodes v to w in JSON format.
func EncodeJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		loggerFromRequest(r).Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
	}
}

func loggerFromRequest(r *http.Request) *zerolog.Logger {
	logger := hlog.FromRequest(r).With().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Logger()
	return &logger
}
