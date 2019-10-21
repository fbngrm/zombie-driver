package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/middleware"
	"github.com/heetch/FabianG-technical-test/types"
	"github.com/rs/zerolog"
)

func newLocationHandler(rf RangeFetcher, logger zerolog.Logger) (http.Handler, error) {
	var mw []middleware.Middleware
	mw = append(mw, middleware.NewRecoverHandler())
	mw = append(mw, middleware.NewContextLog(logger)...)
	mc := middleware.NewMetricsConfig().WithTimeHist(responseTimeHistogram)
	mw = append(mw, middleware.NewMetricsHandler(mc))

	lh := &locationHandler{rf}
	router := mux.NewRouter()
	router.Handle("/drivers/{id:[0-9]+}/locations", middleware.Use(lh, mw...)).Methods("GET").Queries("minutes", "{minutes}")
	router.Handle("/ready", &handler.ReadinessHandler{})
	return router, nil
}

// RangeFetcher provides a method to fetch all the elements in a set at key
// with a score between min and max (including elements with score equal to
// min or max).
type RangeFetcher interface {
	FetchRange(key string, min, max int64) ([]string, error)
}

type locationHandler struct {
	RangeFetcher
}

func (l *locationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	minutes, err := strconv.Atoi(r.FormValue("minutes"))
	if err != nil {
		handler.WriteError(w, r, err, http.StatusBadRequest)
		return
	}
	t := time.Now()
	min := t.Add(-1 * time.Duration(minutes) * time.Minute).UnixNano()
	locations, err := l.FetchRange(id, min, t.UnixNano())
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	var locs []types.LocationUpdate
	for _, s := range locations {
		var l types.LocationUpdate
		err = json.Unmarshal([]byte(s), &l)
		if err != nil {
			handler.WriteError(w, r, err, http.StatusInternalServerError)
			return
		}
		locs = append(locs, l)
	}
	handler.EncodeJSON(w, r, locs, 200)
}
