package server

import "testing"

var distanceTests = []struct {
	lat1  float64
	long1 float64
	lat2  float64
	long2 float64
	dist  float64
}{
	{
		lat1:  53.32055555555556,
		long1: -1.7297222222222221,
		lat2:  53.31861111111111,
		long2: -1.6997222222222223,
		dist:  2.00436783827169,
	},
	{
		lat1:  48.1372,
		long1: 11.5756,
		lat2:  52.5186,
		long2: 13.4083,
		dist:  504.21571518252614,
	},
}

func TestHaversine(t *testing.T) {
	for _, tt := range distanceTests {
		if w, g := tt.dist, haversineKm(tt.lat1, tt.long1, tt.lat2, tt.long2); w != g {
			t.Errorf("want %f but got %f", w, g)
		}
	}
}
