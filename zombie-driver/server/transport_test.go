package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path"
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
		t.Log(dist)
		if w, g := tt.D, dist; w != g {
			t.Errorf("haversine distance: want %f but got %f", w, g)
		}
	}
}

type zombieReq struct {
	d string // description of test case
	p string // URL path of test requests
	m string // HTTP method of test requests
	i string // driver id
	s int    // expected response status code
	r string // expected response
}

type handlerTest struct {
	d     string            // description of test
	p     string            //path of driver location service mock
	r     map[string]string // response payload of driver-location service
	cases []zombieReq
}

// driver-location service response by driver ID
var testLocations = map[string]string{
	"0": "",
	"1": `[
  {
    "latitude": 48.864193,
    "longitude": 2.350498,
    "updated_at": "2018-04-05T22:36:16Z"
  },
  {
    "latitude": 48.863921,
    "longitude":  2.349211,
    "updated_at": "2018-04-05T22:36:21Z"
  }
]`,
	"2": `[
  {
    "latitude": 48.1372,
    "longitude": 11.5756,
    "updated_at": "2018-04-05T22:36:16Z"
  },
  {
    "latitude": 52.5186,
    "longitude":  13.4083,
    "updated_at": "2018-04-05T22:36:21Z"
  }
]`,
}

// FIXME: in a real world scenario we would never hardcode an auth token, not
// even a test token! This should be loaded from e.g. an env file or a secret.
// We would not check-in env files into a remote git repo. It is implemented
// here due to time limitations during coding challenge/prototyping.
var zombieTest = handlerTest{
	d: "test HTTP zombie handler",
	p: "/drivers/%s/locations?minutes=5", //path of driver location service mock
	r: testLocations,
	cases: []zombieReq{
		zombieReq{
			d: "should accept GET only",
			p: "/drivers",
			i: "0",     // empty response
			m: "PATCH", // TODO: should check all HTTP verbs
			s: http.StatusMethodNotAllowed,
		},
		zombieReq{
			d: "expect zombie",
			p: "/drivers",
			i: "1", // zombie true
			m: "GET",
			s: http.StatusOK,
			r: `{"id":1,"zombie":true}`,
		},
		zombieReq{
			d: "expect human",
			p: "/drivers",
			i: "2", // zombie false
			m: "GET",
			s: http.StatusOK,
			r: `{"id":2,"zombie":false}`,
		},
		zombieReq{
			d: "expect StatusNotFound",
			p: "/drivers",
			i: "3", // user not found
			m: "GET",
			s: http.StatusNotFound,
		},
	},
}

func TestProxy(t *testing.T) {
	// mute logger in tests
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)
	driverLocationSrvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: test URL with regex
		segments := strings.Split(r.URL.Path, "/")
		if len(segments) != 4 {
			t.Fatalf("expect 4 path segments but got %d", len(segments))
			w.WriteHeader(http.StatusNotFound)
		}
		id := segments[2]
		if p, ok := zombieTest.r[id]; ok {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(p))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer driverLocationSrvc.Close()

	// create the proxy handler to test
	driverLocationURL := driverLocationSrvc.URL + zombieTest.p
	h, err := newZombieHandler(driverLocationURL, 500, logger)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", zombieTest.d, err)
	}

	zombieService := httptest.NewServer(h)
	defer zombieService.Close()
	zombieClient := zombieService.Client()

	for _, tc := range zombieTest.cases {
		t.Run(tc.d, func(t *testing.T) {
			req, _ := http.NewRequest(tc.m, zombieService.URL+path.Join(tc.p, tc.i), nil)
			req.Close = true // close TCP conn after response was read
			req.Header.Set("Connection", "close")
			res, err := zombieClient.Do(req)
			if err != nil {
				t.Fatalf("%s: unexpected error %v", tc.d, err)
			}
			if w, g := tc.s, res.StatusCode; w != g {
				t.Errorf("%s: expect status code %d got %d", tc.d, w, g)
			}
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("%s: failed to read response %v", tc.d, err)
			}
			if w, g := tc.r, strings.TrimSpace(string(data)); w != g {
				t.Errorf("%s: expect response %s got %s", tc.d, w, g)
			}
		})
	}
}
