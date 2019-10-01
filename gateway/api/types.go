package api

import (
	"fmt"
	"net/http"
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
