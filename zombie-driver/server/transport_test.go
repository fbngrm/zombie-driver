package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/heetch/FabianG-technical-test/testdata"
	"github.com/rs/zerolog"
)

func TestHaversine(t *testing.T) {
	for _, tt := range testdata.Distances {
		var dist float64
		for i := 0; i < len(tt.L)-1; i++ {
			dist += haversineKm(tt.L[i].Lat, tt.L[i].Long, tt.L[i+1].Lat, tt.L[i+1].Long)
		}
		if w, g := tt.D, dist; w != g {
			t.Errorf("haversine distance: want %f but got %f", w, g)
		}
	}
}

// testdata by driver ID and minutes
var zombieTests = map[string]map[int]struct {
	l  string  // mock response of driver-location
	d  string  // description of test case
	p  string  // request path
	r  string  // expected response data
	zr float64 // zombie radius
	s  int     // expected response status code
}{
	"0": { // driver ID 0; test faulty driver-location service
		1: { // 1 minute
			d: "expect InternalServerError when the driver-location mock service is not reachable",
			p: "/drivers/0",
			r: `{"error":"internal_error"}`,
			s: http.StatusInternalServerError,
		},
	},
	"1": { // test empty results
		1: {
			l: "null",
			d: "expect null response from location-service to result in StatusNotFound",
			p: "/drivers/1",
			s: http.StatusNotFound,
		},
	},
	"2": {
		1: { // zombie
			l:  testdata.Drives[0].Loc, // 116.51m
			d:  "expect driver 2 to be a zombie; #1",
			p:  "/drivers/2",
			r:  `{"id":2,"zombie":true}`,
			zr: 400.0,
			s:  http.StatusOK,
		},
		2: { // zombie
			l:  testdata.Drives[1].Loc, // 233.02m
			d:  "expect driver 2 to be a zombie; #2",
			p:  "/drivers/2",
			r:  `{"id":2,"zombie":true}`,
			zr: 400.0,
			s:  http.StatusOK,
		},
		4: { // something else
			l:  testdata.Drives[2].Loc, // 466.04m
			d:  "expect driver 2 to not be a zombie",
			p:  "/drivers/2",
			r:  `{"id":2,"zombie":false}`,
			zr: 400.0,
			s:  http.StatusOK,
		},
	},
	"3": { // test unknown ID
		1: {
			d: "expect StatusNotFound for unknown user",
			p: "/drivers/404",
			s: http.StatusNotFound,
		},
	},
	"4": { // test malformatted json
		1: {
			l: `{"foo":"bar"`,
			d: "expect InternalServerError for unexpected end of JSON input",
			p: "/drivers/4",
			r: `{"error":"internal_error"}`,
			s: http.StatusInternalServerError,
		},
	},
	"5": { // test wrong type of query param id
		1: { // NOTE: type check of ID query param is performed by router only so we expect 404 instead of 500
			l: testdata.Drives[2].Loc, // 466.04m
			d: "expect StatusNotFound for non-int query param ID",
			p: "/drivers/invalidIdType",
			r: "404 page not found",
			s: http.StatusNotFound,
		},
	},
}

func TestProxy(t *testing.T) {
	// mute logger in tests
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)

	// mock backend as target for proxy
	driverLocationSrvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// driver id
		segments := strings.Split(r.URL.Path, "/")
		if len(segments) != 4 {
			t.Fatalf("expect 4 path segments but got %d", len(segments))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		id := segments[2]

		// backend not reachable
		if id == "0" {
			c, _, _ := w.(http.Hijacker).Hijack()
			c.Close()
			return
		}

		// minutes query param
		minutes, ok := r.URL.Query()["minutes"]
		if !ok {
			t.Fatal("expect minutes query parameter")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		m, err := strconv.Atoi(minutes[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
			return
		}
		if tt, ok := zombieTests[id][m]; ok {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(tt.l))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer driverLocationSrvc.Close()

	for id := range zombieTests {
		for minutes := range zombieTests[id] {
			tt := zombieTests[id][minutes]

			// proxy handler to test
			driverLocationURL := driverLocationSrvc.URL + "/drivers/%s/locations?minutes=%d"
			// we use the zombie radius and the minutes of the test data to configure the handler
			h, err := newZombieHandler(driverLocationURL, tt.zr, minutes, logger)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			zombieService := httptest.NewServer(h)
			defer zombieService.Close()
			zombieClient := zombieService.Client()

			t.Run(tt.d, func(t *testing.T) {
				req, err := http.NewRequest("GET", zombieService.URL+tt.p, nil)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				req.Close = true
				req.Header.Set("Connection", "close")
				res, err := zombieClient.Do(req)
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}

				if w, g := tt.s, res.StatusCode; w != g {
					t.Errorf("want status code %d got %d", w, g)
				}
				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("failed to read response %v", err)
				}
				if w, g := tt.r, strings.TrimSpace(string(data)); w != g {
					t.Errorf("want response %s got %s", w, g)
				}
			})
		}
	}
}
