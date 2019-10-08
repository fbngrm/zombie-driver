package handler

import (
	"net/http"
)

type ReadinessHandler struct{}

func (h *ReadinessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(health())
}
