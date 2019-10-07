package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitly/timer_metrics"
	nsq "github.com/nsqio/go-nsq"
)

var (
	topic       = flag.String("topic", "locations", "nsq topic")
	channel     = flag.String("channel", "nsq_to_http", "nsq channel")
	maxInFlight = flag.Int("max-in-flight", 200, "max number of messages to allow in flight")

	numPublishers = flag.Int("n", 100, "number of concurrent publishers")
	sample        = flag.Float64("sample", 1.0, "% of messages to publish (float b/w 0 -> 1)")
	statusEvery   = flag.Int("status-every", 250, "the # of requests between logging status (per handler), 0 disables")

	nsqdTCPAddrs     []string
	lookupdHTTPAddrs []string
)

func init() {
	// flag.Var(nsqdTCPAddrs, "nsqd-tcp-address", "nsqd TCP address (may be given multiple times)")
	// flag.Var(lookupdHTTPAddrs, "lookupd-http-address", "lookupd HTTP address (may be given multiple times)")
	nsqdTCPAddrs = []string{"127.0.0.1:4150"}
	lookupdHTTPAddrs = []string{"127.0.0.1:4161"}
}

type Publisher interface {
	Publish([]byte)
}

type PublishHandler struct {
	// 64bit atomic vars need to be first for proper alignment on 32bit platforms
	counter uint64

	Publisher

	perAddressStatus map[string]*timer_metrics.TimerMetrics
	timermetrics     *timer_metrics.TimerMetrics
}

func (ph *PublishHandler) HandleMessage(m *nsq.Message) error {
	if *sample < 1.0 && rand.Float64() > *sample {
		return nil
	}

	startTime := time.Now()
	ph.Publish(m.Body)
	ph.timermetrics.Status(startTime)

	return nil
}

type PostPublisher struct{}

func (p *PostPublisher) Publish(msg []byte) {
	fmt.Println(string(msg))
}

func main() {
	var publisher Publisher

	cfg := nsq.NewConfig()

	flag.Var(&nsq.ConfigFlag{cfg}, "consumer-opt", "option to passthrough to nsq.Consumer (may be given multiple times, http://godoc.org/github.com/nsqio/go-nsq#Config)")
	flag.Parse()

	if *topic == "" || *channel == "" {
		log.Fatal("--topic and --channel are required")
	}

	if *sample > 1.0 || *sample < 0.0 {
		log.Fatal("ERROR: --sample must be between 0.0 and 1.0")
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	publisher = &PostPublisher{}

	cfg.MaxInFlight = *maxInFlight

	consumer, err := nsq.NewConsumer(*topic, *channel, cfg)
	if err != nil {
		log.Fatal(err)
	}

	perAddressStatus := make(map[string]*timer_metrics.TimerMetrics)

	handler := &PublishHandler{
		Publisher:        publisher,
		perAddressStatus: perAddressStatus,
		timermetrics:     timer_metrics.NewTimerMetrics(*statusEvery, "[aggregate]:"),
	}
	consumer.AddConcurrentHandlers(handler, *numPublishers)

	err = consumer.ConnectToNSQDs(nsqdTCPAddrs)
	if err != nil {
		log.Fatal(err)
	}

	err = consumer.ConnectToNSQLookupds(lookupdHTTPAddrs)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-consumer.StopChan:
			return
		case <-termChan:
			consumer.Stop()
		}
	}
}
