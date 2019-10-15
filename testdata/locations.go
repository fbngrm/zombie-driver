package testdata

// Test data of driving a roundabout in Libreville/Congo for a few minutes.
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
