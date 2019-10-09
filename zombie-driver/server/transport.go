package server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/middleware"
	"github.com/rs/zerolog"
)

func newZombieHandler(driverLocationURL string, logger zerolog.Logger) (http.Handler, error) {
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

	lh := &zombieHandler{
		client: &http.Client{},
		url:    driverLocationURL,
	}

	router := mux.NewRouter()
	router.Handle("/drivers/{id:[0-9]+}", middleware.Use(lh, mw...)).Methods("GET")
	router.Handle("/ready", &handler.ReadinessHandler{})
	return router, nil
}

// Client is a Redis client representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type zombieHandler struct {
	client *http.Client
	url    string
}

func (z *zombieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	request, err := http.NewRequest("GET", fmt.Sprintf(z.url, id), nil)
	if err != nil {
		// do not expose error details, not even for internal endpoints
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	// use auth header from internal request
	request.Header.Set("Authorization", r.Header.Get("Authorization"))
	response, err := z.client.Do(request)
	if err != nil {
		// do not expose error details, not even for internal endpoints
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// do not expose error details, not even for internal endpoints
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	fmt.Println(string(data))
	w.WriteHeader(http.StatusOK)
}
