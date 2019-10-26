package consumer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"sync/atomic"

	nsq "github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
)

// not save for concurrent access; cfg.NumPublishers must be 1
var consumerTests = []struct {
	d string // description of test case
	b []byte // published message body
}{
	{
		d: "expect successful sequential read; #1",
		b: []byte(`{"id":"0","latitude":0.40059538,"longitude":9.43746775}`),
	},
	{
		d: "expect successful sequential read; #2",
		b: []byte(`{"id":"1","latitude":0.40073485,"longitude":9.43776816}`),
	},
	{
		d: "expect successful sequential read; #3",
		b: []byte(`{"id":"2","latitude":0.50073485,"longitude":5.43776816}`),
	},
	{
		d: "expect successful sequential read; #4",
		b: []byte(`{"id":"foo","latitude":0.50073485,"longitude":5.43776816}`),
	},
}

// handler is a mock nsq handler which checks the handles messages body.
type handler struct {
	received uint32
	t        *testing.T
}

func (h *handler) HandleMessage(m *nsq.Message) error {
	i := atomic.LoadUint32(&h.received)
	if w, g := string(consumerTests[i].b), string(m.Body); w != g {
		h.t.Errorf("%s: want %+v got %+v", consumerTests[i].d, w, g)
	}
	atomic.AddUint32(&h.received, 1)
	return nil
}

// test needs to set a timeout
func TestNSQ(t *testing.T) {
	// mute logger
	logger := zerolog.New(ioutil.Discard)
	log.SetFlags(0)
	log.SetOutput(logger)

	// test config
	topic := "consumer-test"
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = 1
	ncfg := &NSQConfig{
		NumPublishers:    1,
		Topic:            topic,
		Channel:          "ch",
		LookupdHTTPAddrs: []string{"127.0.0.1:4161"},
		NsqdTCPAddrs:     []string{"127.0.0.1:4150"},
		Cfg:              cfg,
	}

	// mock handler
	h := &handler{t: t}
	// TODO: mute logger
	consumer, err := NewNSQ(ncfg, h, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// we want to stop consuming once the count of sent and received messages
	// equals. we need to set a test deadline hence this possibly blocks forever
	// in case of failed transmission. we check count every 100ms.
	msgCount := len(consumerTests)
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			if int(atomic.LoadUint32(&h.received)) == msgCount {
				consumer.Shutdown()
				ticker.Stop()
				return
			}
		}
	}()

	for _, tt := range consumerTests {
		err := sendMessage(topic, tt.b)
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
	consumer.Run()
}

// sendMessage is a helper function to send messages to the nsq server.
func sendMessage(topic string, body []byte) error {
	httpclient := &http.Client{}
	endpoint := fmt.Sprintf("http://127.0.0.1:4151/pub?topic=%s", topic)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := httpclient.Do(req)
	if err != nil {
		return err
	}
	if w, g := resp.StatusCode, http.StatusOK; w != g {
		return fmt.Errorf("want status code %d got %d", w, g)
	}
	resp.Body.Close()
	return nil
}
