package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/middleware"
	"github.com/rs/zerolog"
)

func newLocationHandler(redisAddr string, logger zerolog.Logger) (http.Handler, error) {
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

	lh := &locationHandler{
		c: redis.NewClient(&redis.Options{Addr: redisAddr}),
	}
	router := mux.NewRouter()
	router.Handle("/drivers/{id:[0-9]+}/locations", middleware.Use(lh, mw...)).Methods("GET").Queries("minutes", "{minutes}")
	router.Handle("/ready", &handler.ReadinessHandler{})
	return router, nil
}

// Client is a Redis client representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type locationHandler struct {
	c *redis.Client // thread safe
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
	opt := redis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(t.Unix(), 10),
	}
	locations := l.c.ZRangeByScore(id, &opt)
	if err := locations.Err(); err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	var locs []LocationUpdate
	for _, s := range locations.Val() {
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
