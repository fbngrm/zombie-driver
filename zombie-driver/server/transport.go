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
	var mw []middleware.Middleware
	mw = append(mw, middleware.NewRecoverHandler())
	mw = append(mw, middleware.NewContextLog(logger)...)
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

// ServeHTTP fetches location updates from the drive-ocation service and 
// determines if the given driver ID identifies a zombie. If are no locations udpates available, we do not assume the driver is a zombie. If there are updates available, the driver is considered to be a zombie if the total distance he moved during the 
func (z *zombieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// fetch locations from driver-location service
	request, err := http.NewRequest("GET", fmt.Sprintf(z.url, id), nil)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	response, err := z.client.Do(request)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	// something wrong with the URL
	if response.StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	// we rely on locations being sorted by update time; either de- or ascending
	var locs []LocationUpdate
	if err = json.Unmarshal(data, &locs); err != nil {
		handler.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	// no data found for the driver ID
	if len(locs) == 0 {
		w.WriteHeader(http.StatusNotFound)
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

const degreesToRadians = math.Pi / 180.0

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
