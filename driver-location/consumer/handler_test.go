package consumer

import (
	"regexp"
	"testing"

	"github.com/heetch/FabianG-technical-test/types"
	nsq "github.com/nsqio/go-nsq"
)

// matches dates in RFC339 format
const rfc3339 = `^([\d]+)-(0[1-9]|1[012])-(0[1-9]|[12][\d]|3[01])[Tt]([01][\d]|2[0-3]):([0-5][\d]):([0-5][\d]|60)(\.[\d]+)?(([Zz])|([\+|\-]([01][\d]|2[0-3]):[0-5][\d]))$`

var tests = map[string]struct {
	d string               // test case description
	m *nsq.Message         // input
	l types.LocationUpdate // expected output
}{
	"1": {
		d: "expect LocationUpdates to equal; #1",
		m: nsq.NewMessage(
			nsq.MessageID{},
			[]byte(`{"id":"1","latitude":0.40059538,"longitude":9.43746775}`)),
		l: types.LocationUpdate{
			Lat:  0.40059538,
			Long: 9.43746775,
		},
	},
	"2": {
		d: "expect LocationUpdates to equal; #2",
		m: nsq.NewMessage(
			nsq.MessageID{},
			[]byte(`{"id":"2","latitude":0.50059538,"longitude":9.53746775}`)),
		l: types.LocationUpdate{
			Lat:  0.50059538,
			Long: 9.53746775,
		},
	},
}

// testPublisher mocks a Publisher. It checks for message equality
// and timestamp format.
type testPublisher struct {
	t *testing.T
}

func (t *testPublisher) Publish(timestamp int64, key string, l types.LocationUpdate) error {
	w, g := tests[key].l, l
	if !(w.Lat == g.Lat && w.Long == w.Long) {
		t.t.Errorf("%s: want %v got %v", tests[key].d, w, g)
	}
	r, err := regexp.Compile(rfc3339)
	if err != nil {
		t.t.Fatalf("unexpected error: %v", err)
	}
	if w, g := 1, len(r.FindAllString(l.UpdatedAt, -1)); w != g {
		t.t.Errorf("%s: want %d match got %d", tests[key].d, w, g)
	}
	return nil
}

func TestLocationUpdater(t *testing.T) {
	p := testPublisher{
		t: t,
	}
	lu := LocationUpdater{
		&p,
	}
	for key := range tests {
		tt := tests[key]
		lu.HandleMessage(tt.m)
	}
}
