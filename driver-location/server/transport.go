package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/gateway/api/middleware"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/rs/zerolog"
)

func newLocationHandler(logger zerolog.Logger) (http.Handler, error) {
	// initialize middleware common to all handlers
	var mw []middleware.Middleware
	mw = append(mw,
		middleware.NewAuthCheck(authtoken),
		middleware.NewRecoverHandler(),
	)
	mw = append(mw, middleware.NewContextLog(logger)...)
	// we measure response time only for all handlers
	mc := middleware.NewMetricsConfig().WithTimeHist(responseTimeHistogram)
	mw = append(mw, middleware.NewMetricsHandler(mc))

	router := mux.NewRouter()
	router.Handle("/drivers/{id:[0-9]+}/locations", middleware.Use(&locationHandler{}, mw...)).Methods("GET").Queries("minutes")
	router.Handle("/ready", &handler.ReadinessHandler{})
	return router, nil
}

type locationHandler struct{}

func (l *locationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
