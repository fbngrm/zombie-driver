package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/middleware"
	"github.com/heetch/FabianG-technical-test/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	responseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "driver_location_response_time",
			Help:    "histogram of response times for driver-location http handler",
			Buckets: prometheus.ExponentialBuckets(0.5e-3, 2, 14), // 0.5ms to 4s
		},
		[]string{"path", "status_code"},
	)
)

func init() {
	prometheus.MustRegister(responseTimeHistogram)
}

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

// locationHandler respondes to driver location requests.
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

	var locations []string
	if err := hystrix.Do("fetch_redis", func() error { // circuit-breaker
		locations, err = l.FetchRange(id, min, t.UnixNano())
		return err
	}, nil); err != nil {
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
