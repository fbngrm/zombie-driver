package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/driver-location/cmd/cli"
	"github.com/heetch/FabianG-technical-test/driver-location/consumer"
	"github.com/heetch/FabianG-technical-test/driver-location/server"
	"github.com/heetch/FabianG-technical-test/metrics"
	nsq "github.com/nsqio/go-nsq"
)

var (
	service     = "driver-location"
	httpAddr    = ":8081"
	redisAddr   = ":6379"
	metricsAddr = ":9103"

	nsqdTCPAddrs     = []string{"localhost:4150"}
	lookupdHTTPAddrs = []string{"localhost:4161"}
	topic            = "locations"
	channel          = "loc-chan"
	numPublishers    = 100
	maxInflight      = 250
)

func main() {
	logger := cli.NewLogger(service)

	httpSrv, err := server.New(httpAddr, redisAddr, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s service: %v\n", service, err)
		os.Exit(2)
	}

	metricsSrv := metrics.NewMetrics(metricsAddr, logger)

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = maxInflight
	ncfg := &consumer.NSQConfig{
		NumPublishers:    numPublishers,
		Topic:            topic,
		Channel:          channel,
		RedisAddr:        redisAddr,
		LookupdHTTPAddrs: lookupdHTTPAddrs,
		NsqdTCPAddrs:     nsqdTCPAddrs,
		Cfg:              cfg,
	}
	nsqConsumer, err := consumer.NewNSQ(ncfg, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s service: %v\n", service, err)
		os.Exit(2)
	}

	cli.RunServer(httpSrv, metricsSrv, nsqConsumer)
}
