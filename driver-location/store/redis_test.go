package store

import (
	"reflect"
	"testing"

	"github.com/go-redis/redis"
	"github.com/heetch/FabianG-technical-test/types"
)

// publish tests by driver-ID
var publishTests = map[string]struct {
	d string               // test case description
	t int64                // input timestamp
	l types.LocationUpdate // input
	k string               // expected key
	m redis.Z              // expected Z
}{
	"0": {
		d: "expect success; #1",
		t: 1257894000,
		l: types.LocationUpdate{
			UpdatedAt: "2019-10-15T07:00:07Z",
			Lat:       0.40059538,
			Long:      9.43746775,
		},
		k: "0",
		m: redis.Z{
			Score:  float64(1257894000),
			Member: `{"updated_at":"2019-10-15T07:00:07Z","latitude":0.40059538,"longitude":9.43746775}`,
		},
	},
	"1": {
		d: "expect success; #2",
		t: 1257895000,
		l: types.LocationUpdate{
			UpdatedAt: "2019-10-15T07:00:07Z",
			Lat:       0.50059538,
			Long:      9.53746775,
		},
		k: "1",
		m: redis.Z{
			Score:  float64(1257895000),
			Member: `{"updated_at":"2019-10-15T07:00:07Z","latitude":0.50059538,"longitude":9.53746775}`,
		},
	},
}

// range tests by driver-ID
var rangeTests = map[string]struct {
	d   string         // test case description
	min int64          // input min score
	max int64          // input max score
	k   string         // expected key
	z   redis.ZRangeBy // expected range
}{
	"0": {
		d:   "expect success; #1",
		min: 1257895000,
		max: 1257891000,
		k:   "0",
		z: redis.ZRangeBy{
			Min: "1257895000",
			Max: "1257891000",
		},
	},
	"1": {
		d:   "expect success; #2",
		min: 1257896000,
		max: 1257896001,
		k:   "1",
		z: redis.ZRangeBy{
			Min: "1257896000",
			Max: "1257896001",
		},
	},
}

type testRedis struct {
	t *testing.T
}

func (r *testRedis) ZAddNX(key string, member redis.Z) error {
	if w, g := publishTests[key].k, key; w != g {
		r.t.Errorf("%s: want %s got %s", publishTests[key].d, w, g)
	}
	if w, g := publishTests[key].m, member; !reflect.DeepEqual(w, g) {
		r.t.Errorf("%s: want %+v got %+v", publishTests[key].d, w, g)
	}
	return nil
}

func (r *testRedis) ZRangeByScore(key string, opt redis.ZRangeBy) ([]string, error) {
	if w, g := rangeTests[key].k, key; w != g {
		r.t.Errorf("%s: want %s got %s", publishTests[key].d, w, g)
	}
	if w, g := rangeTests[key].z, opt; !reflect.DeepEqual(w, g) {
		r.t.Errorf("%s: want %+v got %+v", publishTests[key].d, w, g)
	}
	return nil, nil
}

func TestPublish(t *testing.T) {
	r := Redis{
		&testRedis{t: t},
	}
	for k, tt := range publishTests {
		err := r.Publish(tt.t, k, tt.l)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestFetchRange(t *testing.T) {
	r := Redis{
		&testRedis{t: t},
	}
	for k, tt := range rangeTests {
		_, err := r.FetchRange(k, tt.min, tt.max)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}
