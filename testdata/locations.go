package testdata

// This module contains test data of driving a roundabout in Libreville/Congo
// for a few minutes.

// GPS location
type Location struct {
	Lat  float64
	Long float64
}

var Locations = []Location{
	{
		Lat:  0.40059538,
		Long: 9.43746775,
	},
	{ // 36.83m distance to previous location
		Lat:  0.40073485,
		Long: 9.43776816,
	},
	{ // 19.68m
		Lat:  0.40091187,
		Long: 9.43776816,
	},
	{ // 13.72m
		Lat:  0.40091187,
		Long: 9.43764478,
	},
	{ // 23.02m
		Lat:  0.40080459,
		Long: 9.43746775,
	},
	{ // 23.26m
		Lat:  0.40059538,
		Long: 9.43746775,
	},
}

type LocationDist struct {
	L []Location
	D float64 // total distance between all locations in km
}

var Distances = []LocationDist{
	{
		L: []Location{
			Locations[0],
			Locations[1],
		},
		D: 0.03682779259699572,
	},
	{
		L: []Location{
			Locations[1],
			Locations[2],
		},
		D: 0.01968372591462287,
	},
	{
		L: []Location{
			Locations[2],
			Locations[3],
		},
		D: 0.013718894195373406,
	},
	{
		L: []Location{
			Locations[3],
			Locations[4],
		},
		D: 0.023016835548780985,
	},
	{
		L: []Location{
			Locations[4],
			Locations[5],
		},
		D: 0.023263090603309826,
	},
	{
		L: []Location{
			Locations[5],
			Locations[0],
		},
		D: 0,
	},
	{
		L: []Location{
			Locations[0],
			Locations[1],
			Locations[2],
		},
		D: 0.05651151851161859,
	},
	{
		L: []Location{
			Locations[0],
			Locations[1],
			Locations[2],
			Locations[3],
		},
		D: 0.070230412706992,
	},
	{
		L: []Location{
			Locations[0],
			Locations[1],
			Locations[2],
			Locations[3],
			Locations[4],
		},
		D: 0.09324724825577299,
	},
	{
		L: []Location{
			Locations[0],
			Locations[1],
			Locations[2],
			Locations[3],
			Locations[4],
			Locations[5],
		},
		D: 0.11651033885908282,
	},
	{
		L: []Location{
			Locations[0],
			Locations[1],
			Locations[2],
			Locations[3],
			Locations[4],
			Locations[5],
			Locations[0],
		},
		D: 0.11651033885908282,
	},
}
var Drives = []struct {
	Locs     []string // location updates as json lines
	Loc      string   // array of location updates as a single string
	Dist     float64  // distance in meters
	Mins     int64    // minutes of driving
	Interval int      // time interval of location updates in seconds
}{
	{ // one round 116.51m; update interval 10s
		Locs: []string{
			`{"updated_at":"2019-10-15T07:00:07Z","latitude":0.40059538,"longitude":9.43746775}`,
			`{"updated_at":"2019-10-15T07:10:07Z","latitude":0.40073485,"longitude":9.43776816}`, // 36.83m
			`{"updated_at":"2019-10-15T07:20:07Z","latitude":0.40091187,"longitude":9.43776816}`, // 19.68m
			`{"updated_at":"2019-10-15T07:30:07Z","latitude":0.40091187,"longitude":9.43764478}`, // 13.72m
			`{"updated_at":"2019-10-15T07:40:07Z","latitude":0.40080459,"longitude":9.43746775}`, // 23.02m
			`{"updated_at":"2019-10-15T07:50:07Z","latitude":0.40059538,"longitude":9.43746775}`, // 23.26m
		},
		Loc:      `[{"updated_at":"2019-10-15T07:00:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T07:10:07Z","latitude":0.40073485,"longitude":9.43776816},{"updated_at":"2019-10-15T07:20:07Z","latitude":0.40091187,"longitude":9.43776816},{"updated_at":"2019-10-15T07:30:07Z","latitude":0.40091187,"longitude":9.43764478},{"updated_at":"2019-10-15T07:40:07Z","latitude":0.40080459,"longitude":9.43746775},{"updated_at":"2019-10-15T07:50:07Z","latitude":0.40059538,"longitude":9.43746775}]`,
		Dist:     116.51,
		Mins:     1,
		Interval: 10,
	},
	{
		Locs: []string{ // two rounds 233.02m; update interval 10s
			`{"updated_at":"2019-10-15T07:00:07Z","latitude":0.40059538,"longitude":9.43746775}`,
			`{"updated_at":"2019-10-15T07:10:07Z","latitude":0.40073485,"longitude":9.43776816}`, // 36.83m
			`{"updated_at":"2019-10-15T07:20:07Z","latitude":0.40091187,"longitude":9.43776816}`, // 19.68m
			`{"updated_at":"2019-10-15T07:30:07Z","latitude":0.40091187,"longitude":9.43764478}`, // 13.72m
			`{"updated_at":"2019-10-15T07:40:07Z","latitude":0.40080459,"longitude":9.43746775}`, // 23.02m
			`{"updated_at":"2019-10-15T07:50:07Z","latitude":0.40059538,"longitude":9.43746775}`, // 23.26m
			`{"updated_at":"2019-10-15T08:00:07Z","latitude":0.40059538,"longitude":9.43746775}`, // pause
			`{"updated_at":"2019-10-15T08:10:07Z","latitude":0.40073485,"longitude":9.43776816}`, // 36.83m
			`{"updated_at":"2019-10-15T08:20:07Z","latitude":0.40091187,"longitude":9.43776816}`, // 19.68m
			`{"updated_at":"2019-10-15T08:30:07Z","latitude":0.40091187,"longitude":9.43764478}`, // 13.72m
			`{"updated_at":"2019-10-15T08:40:07Z","latitude":0.40080459,"longitude":9.43746775}`, // 23.02m
			`{"updated_at":"2019-10-15T08:50:07Z","latitude":0.40059538,"longitude":9.43746775}`, // 23.26m
		},
		Loc:      `[{"updated_at":"2019-10-15T07:00:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T07:10:07Z","latitude":0.40073485,"longitude":9.43776816},{"updated_at":"2019-10-15T07:20:07Z","latitude":0.40091187,"longitude":9.43776816},{"updated_at":"2019-10-15T07:30:07Z","latitude":0.40091187,"longitude":9.43764478},{"updated_at":"2019-10-15T07:40:07Z","latitude":0.40080459,"longitude":9.43746775},{"updated_at":"2019-10-15T07:50:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T08:00:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T08:10:07Z","latitude":0.40073485,"longitude":9.43776816},{"updated_at":"2019-10-15T08:20:07Z","latitude":0.40091187,"longitude":9.43776816},{"updated_at":"2019-10-15T08:30:07Z","latitude":0.40091187,"longitude":9.43764478},{"updated_at":"2019-10-15T08:40:07Z","latitude":0.40080459,"longitude":9.43746775},{"updated_at":"2019-10-15T08:50:07Z","latitude":0.40059538,"longitude":9.43746775}]`,
		Dist:     233.02,
		Mins:     2,
		Interval: 10,
	},
	{
		Locs: []string{ // four rounds 466.04m; update interval 10s
			`{"updated_at":"2019-10-15T07:00:07Z","latitude":0.40059538,"longitude":9.43746775}`,
			`{"updated_at":"2019-10-15T07:10:07Z","latitude":0.40073485,"longitude":9.43776816}`, // 36.83m
			`{"updated_at":"2019-10-15T07:20:07Z","latitude":0.40091187,"longitude":9.43776816}`, // 19.68m
			`{"updated_at":"2019-10-15T07:30:07Z","latitude":0.40091187,"longitude":9.43764478}`, // 13.72m
			`{"updated_at":"2019-10-15T07:40:07Z","latitude":0.40080459,"longitude":9.43746775}`, // 23.02m
			`{"updated_at":"2019-10-15T07:50:07Z","latitude":0.40059538,"longitude":9.43746775}`, // 23.26m
			`{"updated_at":"2019-10-15T08:00:07Z","latitude":0.40059538,"longitude":9.43746775}`, // pause
			`{"updated_at":"2019-10-15T08:10:07Z","latitude":0.40073485,"longitude":9.43776816}`, // 36.83m
			`{"updated_at":"2019-10-15T08:20:07Z","latitude":0.40091187,"longitude":9.43776816}`, // 19.68m
			`{"updated_at":"2019-10-15T08:30:07Z","latitude":0.40091187,"longitude":9.43764478}`, // 13.72m
			`{"updated_at":"2019-10-15T08:40:07Z","latitude":0.40080459,"longitude":9.43746775}`, // 23.02m
			`{"updated_at":"2019-10-15T08:50:07Z","latitude":0.40059538,"longitude":9.43746775}`, // 23.26m
			`{"updated_at":"2019-10-15T09:00:07Z","latitude":0.40059538,"longitude":9.43746775}`, // pause
			`{"updated_at":"2019-10-15T09:10:07Z","latitude":0.40073485,"longitude":9.43776816}`, // 36.83m
			`{"updated_at":"2019-10-15T09:20:07Z","latitude":0.40091187,"longitude":9.43776816}`, // 19.68m
			`{"updated_at":"2019-10-15T09:30:07Z","latitude":0.40091187,"longitude":9.43764478}`, // 13.72m
			`{"updated_at":"2019-10-15T09:40:07Z","latitude":0.40080459,"longitude":9.43746775}`, // 23.02m
			`{"updated_at":"2019-10-15T09:50:07Z","latitude":0.40059538,"longitude":9.43746775}`, // 23.26m
			`{"updated_at":"2019-10-15T10:00:07Z","latitude":0.40059538,"longitude":9.43746775}`, // pause
			`{"updated_at":"2019-10-15T10:10:07Z","latitude":0.40073485,"longitude":9.43776816}`, // 36.83m
			`{"updated_at":"2019-10-15T10:20:07Z","latitude":0.40091187,"longitude":9.43776816}`, // 19.68m
			`{"updated_at":"2019-10-15T10:30:07Z","latitude":0.40091187,"longitude":9.43764478}`, // 13.72m
			`{"updated_at":"2019-10-15T10:40:07Z","latitude":0.40080459,"longitude":9.43746775}`, // 23.02m
			`{"updated_at":"2019-10-15T10:50:07Z","latitude":0.40059538,"longitude":9.43746775}`, // 23.26m
		},
		Loc:      `[{"updated_at":"2019-10-15T07:00:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T07:10:07Z","latitude":0.40073485,"longitude":9.43776816},{"updated_at":"2019-10-15T07:20:07Z","latitude":0.40091187,"longitude":9.43776816},{"updated_at":"2019-10-15T07:30:07Z","latitude":0.40091187,"longitude":9.43764478},{"updated_at":"2019-10-15T07:40:07Z","latitude":0.40080459,"longitude":9.43746775},{"updated_at":"2019-10-15T07:50:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T08:00:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T08:10:07Z","latitude":0.40073485,"longitude":9.43776816},{"updated_at":"2019-10-15T08:20:07Z","latitude":0.40091187,"longitude":9.43776816},{"updated_at":"2019-10-15T08:30:07Z","latitude":0.40091187,"longitude":9.43764478},{"updated_at":"2019-10-15T08:40:07Z","latitude":0.40080459,"longitude":9.43746775},{"updated_at":"2019-10-15T08:50:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T09:00:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T09:10:07Z","latitude":0.40073485,"longitude":9.43776816},{"updated_at":"2019-10-15T09:20:07Z","latitude":0.40091187,"longitude":9.43776816},{"updated_at":"2019-10-15T09:30:07Z","latitude":0.40091187,"longitude":9.43764478},{"updated_at":"2019-10-15T09:40:07Z","latitude":0.40080459,"longitude":9.43746775},{"updated_at":"2019-10-15T09:50:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T10:00:07Z","latitude":0.40059538,"longitude":9.43746775},{"updated_at":"2019-10-15T10:10:07Z","latitude":0.40073485,"longitude":9.43776816},{"updated_at":"2019-10-15T10:20:07Z","latitude":0.40091187,"longitude":9.43776816},{"updated_at":"2019-10-15T10:30:07Z","latitude":0.40091187,"longitude":9.43764478},{"updated_at":"2019-10-15T10:40:07Z","latitude":0.40080459,"longitude":9.43746775},{"updated_at":"2019-10-15T10:50:07Z","latitude":0.40059538,"longitude":9.43746775}]`,
		Dist:     466.04,
		Mins:     4,
		Interval: 10,
	},
}
