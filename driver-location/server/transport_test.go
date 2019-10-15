package server

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heetch/FabianG-technical-test/testdata"
	"github.com/rs/zerolog"
)

type redisTestClient struct{}

func (r *redisTestClient) ZRangeByScore(key string, min, max int64) ([]string, error) {
	mins := (max - min) / 60 // minutes
	return locationTests[key][mins].l, locationTests[key][mins].e
}

var testErr = errors.New("redis_test_error")

// location/error by driver ID and minutes
var locationTests = map[string]map[int64]struct {
	l []string // input locations
	d string   // description of test case
	p string   // request path
	r string   // expected response data
	e error    // expected error
	s int      // expected response status code
}{
	"0": { // driver ID 0; test empty results
		1: { // 1 minute
			l: []string{},
			d: "expect empty redis lookup to result in null response",
			p: "/drivers/0/locations?minutes=1",
			r: "null",
			e: nil,
			s: http.StatusOK,
		},
	},
	"1": { // driver ID 1; test success
		1: { // 1 minute
			l: testdata.Drives[0].Locs,
			d: "expect successful result of 6 locations",
			p: "/drivers/1/locations?minutes=1",
			r: testdata.Drives[0].Loc,
			e: nil,
			s: http.StatusOK,
		},
		2: { // 2 minutes
			l: testdata.Drives[1].Locs,
			d: "expect successful result of 12 locations; #1",
			p: "/drivers/1/locations?minutes=2",
			r: testdata.Drives[1].Loc,
			e: nil,
			s: http.StatusOK,
		},
		3: { // 3 minutes
			l: testdata.Drives[1].Locs,
			d: "expect successful result of 12 locations; #2",
			p: "/drivers/1/locations?minutes=3",
			r: testdata.Drives[1].Loc,
			e: nil,
			s: http.StatusOK,
		},
		4: { // 4 minutes
			l: testdata.Drives[2].Locs,
			d: "expect successful result of 24 locations",
			p: "/drivers/1/locations?minutes=4",
			r: testdata.Drives[2].Loc,
			e: nil,
			s: http.StatusOK,
		},
	},
	"2": { // driver ID 2; error tests
		1: { // 1 minute
			l: nil,
			d: "expect error in redis lookup",
			p: "/drivers/2/locations?minutes=1",
			r: "",
			e: testErr,
			s: http.StatusInternalServerError,
		},
		2: { // 2 minutes
			l: nil,
			d: "expect unmarshalling error",
			p: "/drivers/2/locations?minutes=2",
			r: "",
			e: testErr,
			s: http.StatusInternalServerError,
		},
	},
}

func TestServeHTTP(t *testing.T) {
	// mute logger in tests
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)

	// create the handler to test
	h, err := newLocationHandler(&redisTestClient{}, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// test server
	locationService := httptest.NewServer(h)
	defer locationService.Close()
	locationClient := locationService.Client()

	for id := range locationTests {
		for minutes := range locationTests[id] {
			tt := locationTests[id][minutes]
			t.Run(tt.d, func(t *testing.T) {
				req, _ := http.NewRequest("GET", locationService.URL+tt.p, nil)
				req.Close = true // close TCP conn after response was read
				req.Header.Set("Connection", "close")
				res, err := locationClient.Do(req)
				if err != nil {
					t.Fatalf("%s: unexpected error %v", tt.d, err)
				}
				if w, g := tt.s, res.StatusCode; w != g {
					t.Errorf("%s: expect status code %d got %d", tt.d, w, g)
				}
				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Errorf("%s: failed to read response %v", tt.d, err)
				}
				if w, g := tt.r, strings.TrimSpace(string(data)); w != g {
					t.Errorf("%s: expect response\n%s\ngot\n%s", tt.d, w, g)
				}
			})
		}
	}
}
