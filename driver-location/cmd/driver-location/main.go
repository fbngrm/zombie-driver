package main

import (
	"fmt"
	"os"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/heetch/FabianG-technical-test/driver-location/cmd/driver-location/cli"
	"github.com/heetch/FabianG-technical-test/driver-location/consumer"
	"github.com/heetch/FabianG-technical-test/driver-location/server"
	"github.com/heetch/FabianG-technical-test/driver-location/store"
	"github.com/heetch/FabianG-technical-test/metrics"
	nsq "github.com/nsqio/go-nsq"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unkown"

	service     = kingpin.Flag("service", "service name").Envar("SERVICE").Default("driver-location").String()
	httpAddr    = kingpin.Flag("http-addr", "address of HTTP server").Envar("HTTP_ADDR").Default(":8081").String()
	redisAddr   = kingpin.Flag("redis-addr", "address of Redis instance to connect").Envar("REDIS_ADDR").Default(":6379").String()
	metricsAddr = kingpin.Flag("metrics-addr", "address of metrics server").Envar("METRICS_ADDR").Default(":9103").String()

	// NSQ
	nsqdTCPAddrs     = kingpin.Flag("nsqd-addr", "TCP addresses of NSQ deamon").Envar("NSQD_ADDR").Default(":4150").Strings()
	lookupdHTTPAddrs = kingpin.Flag("lookupd-http-addr", "HTTP addresses for NSQD lookup").Envar("LOOKUPD_HTTP_ADDRS").Default(":4161").Strings()
	topic            = kingpin.Flag("topic", "NSQ topic").Envar("TOPIC").Default("locations").String()
	channel          = kingpin.Flag("channel", "NSQ channel").Envar("CHANNEL").Default("loc-chan").String()
	numPublishers    = kingpin.Flag("num-publishers", "NSQ publishers").Envar("NUM_PUBLISHERS").Default("100").Int()
	maxInflight      = kingpin.Flag("max-inflight", "NSQ max inflight").Envar("MAX_INFLIGHT").Default("250").Int()

	// should be greater then prometheus scrape interval (default 30s); decreased in coding challenge
	shutdownDelay = kingpin.Flag("shutdown-delay", "shutdown delay").Envar("SHUTDOWN_DELAY").Default("5000").Int()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	// configure circuit-breaker
	hystrix.ConfigureCommand("fetch_redis", hystrix.CommandConfig{
		Timeout:               1000, // ms
		MaxConcurrentRequests: 1000,
		ErrorPercentThreshold: 25,
	})
	hystrix.ConfigureCommand("handle_nsq_msg", hystrix.CommandConfig{
		Timeout:               1000, // ms
		MaxConcurrentRequests: 1000,
		ErrorPercentThreshold: 25,
	})

	logger := cli.NewLogger(*service, version)

	redisStore := store.NewRedis(*redisAddr)
	httpSrv, err := server.New(*httpAddr, redisStore, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s service: %v\n", *service, err)
		os.Exit(2)
	}

	metricsSrv := metrics.New(*metricsAddr, logger)

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = *maxInflight
	ncfg := &consumer.NSQConfig{
		NumPublishers:    *numPublishers,
		Topic:            *topic,
		Channel:          *channel,
		LookupdHTTPAddrs: *lookupdHTTPAddrs,
		NsqdTCPAddrs:     *nsqdTCPAddrs,
		Cfg:              cfg,
	}
	handler := &consumer.LocationUpdater{
		redisStore,
	}
	nsqConsumer, err := consumer.NewNSQ(ncfg, handler, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s service: %v\n", *service, err)
		os.Exit(2)
	}

	cli.RunServer(httpSrv, metricsSrv, nsqConsumer, *shutdownDelay)
}
