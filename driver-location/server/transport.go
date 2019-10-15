package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/middleware"
	"github.com/rs/zerolog"
)

type zRangeByScorer interface {
	ZRangeByScore(key string, min, max int64) ([]string, error)
}

func newLocationHandler(z zRangeByScorer, logger zerolog.Logger) (http.Handler, error) {
	var mw []middleware.Middleware
	mw = append(mw, middleware.NewRecoverHandler())
	mw = append(mw, middleware.NewContextLog(logger)...)
	mc := middleware.NewMetricsConfig().WithTimeHist(responseTimeHistogram)
	mw = append(mw, middleware.NewMetricsHandler(mc))

	lh := &locationHandler{z}
	router := mux.NewRouter()
	router.Handle("/drivers/{id:[0-9]+}/locations", middleware.Use(lh, mw...)).Methods("GET").Queries("minutes", "{minutes}")
	router.Handle("/ready", &handler.ReadinessHandler{})
	return router, nil
}

type locationHandler struct {
	zRangeByScorer
}

func (l *locationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	minutes, err := strconv.Atoi(r.FormValue("minutes"))
	if err != nil {
		handler.WriteError(w, r, err, http.StatusBadRequest)
		return
	}
	t := time.Now()
	min := t.Add(-1 * time.Duration(minutes) * time.Minute).Unix()
	locations, err := l.ZRangeByScore(id, min, t.Unix())
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	var locs []LocationUpdate
	for _, s := range locations {
		var l LocationUpdate
		err = json.Unmarshal([]byte(s), &l)
		if err != nil {
			handler.WriteError(w, r, err, http.StatusInternalServerError)
			return
		}
		locs = append(locs, l)
	}
	handler.EncodeJSON(w, r, locs, 200)
}
