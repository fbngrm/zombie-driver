package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Copied from https://github.com/heetch/regula/blob/master/api/types.go
type Error struct {
	Err      string         `json:"error"`
	Response *http.Response `json:"-"` // Will not be marshalled
}

func (e Error) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		e.Response.Request.Method,
		e.Response.Request.URL,
		e.Response.StatusCode,
		e.Err)
}

// HTTP errors
var (
	errInternal   = errors.New("internal_error")
	errBadRequest = errors.New("bad_request")
)

// WriteError writes an error to the http response in JSON format.
func WriteError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// Prepare log.
	logger := LoggerFromRequest(r).With().
		Err(err).
		Int("status", code).
		Logger()

	// Hide error from client if it's internal.
	switch code {
	case http.StatusInternalServerError:
		logger.Error().Msg("unexpected http error")
		err = errInternal
	case http.StatusBadRequest:
		logger.Error().Msg("http error bad request")
		err = errBadRequest
	default:
		logger.Debug().Msg("http error")
	}
	EncodeJSON(w, r, &Error{Err: err.Error()}, code)
}

// EncodeJSON encodes v to w in JSON format.
func EncodeJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		LoggerFromRequest(r).Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
	}
}

func LoggerFromRequest(r *http.Request) *zerolog.Logger {
	logger := hlog.FromRequest(r).With().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Logger()
	return &logger
}
