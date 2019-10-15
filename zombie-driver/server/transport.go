package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/heetch/FabianG-technical-test/handler"
	"github.com/heetch/FabianG-technical-test/middleware"
	"github.com/rs/zerolog"
)

func newZombieHandler(driverLocationURL string, zombieRadius float64, logger zerolog.Logger) (http.Handler, error) {
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
		client:       &http.Client{},
		url:          driverLocationURL,
		zombieRadius: zombieRadius,
	}

	router := mux.NewRouter()
	router.Handle("/drivers/{id:[0-9]+}", middleware.Use(lh, mw...)).Methods("GET")
	router.Handle("/ready", &handler.ReadinessHandler{})
	return router, nil
}

type zombieHandler struct {
	client       *http.Client
	url          string
	zombieRadius float64
}

func (z *zombieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// fetch location updates from location service; if there are no location
	// updates available for the driver ID, we assume it's a zombie. we do not
	// distinguish between non-existing drivers and zombies.
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
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	// user or URL not found
	if len(data) == 0 || response.StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// we rely on locations being sorted by update time
	// TODO: check order
	var locs []LocationUpdate
	err = json.Unmarshal(data, &locs)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	var dist float64
	for i := 0; i < len(locs)-1; i++ {
		dist += haversineKm(locs[i].Lat, locs[i].Long, locs[i+1].Lat, locs[i+1].Long)
	}
	driverId, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	zombie := ZombieDriver{
		ID:     driverId,
		Zombie: dist < z.zombieRadius,
	}
	handler.EncodeJSON(w, r, zombie, http.StatusOK)
}

var degreesToRadians = math.Pi / 180.0

// calculate haversine distance for linear distance
func haversineKm(lat1, long1, lat2, long2 float64) float64 {
	earthRadiusKm := 6371.0

	dlong := (long2 - long1) * degreesToRadians
	dlat := (lat2 - lat1) * degreesToRadians

	lat1 = lat1 * degreesToRadians
	lat2 = lat2 * degreesToRadians

	a := math.Pow(math.Sin(dlat/2.0), 2) +
		math.Pow(math.Sin(dlong/2.0), 2)*math.Cos(lat1)*math.Cos(lat2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}
