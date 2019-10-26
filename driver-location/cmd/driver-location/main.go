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
	httpAddr    = kingpin.Flag("http-addr", "address of HTTP server").Envar("HTTP_ADDR").Required().String()
	metricsAddr = kingpin.Flag("metrics-addr", "address of metrics server").Envar("METRICS_ADDR").Required().String()

	// Redis
	redisAddr = kingpin.Flag("redis-addr", "address of Redis instance to connect").Envar("REDIS_ADDR").Required().String()

	// NSQ
	nsqdTCPAddrs        = kingpin.Flag("nsqd-tcp-addrs", "TCP addresses of NSQ deamon").Envar("NSQD_TCP_ADDRS").Required().Strings()
	nsqLookupdHTTPAddrs = kingpin.Flag("nsqd-lookupd-http-addrs", "HTTP addresses for NSQD lookup").Envar("NSQ_LOOKUPD_HTTP_ADDRS").Required().Strings()
	nsqTopic            = kingpin.Flag("nsqd-topic", "NSQ topic").Envar("NSQ_TOPIC").Required().String()
	nsqChan             = kingpin.Flag("nsqd-chan", "NSQ channel").Envar("NSQ_CHAN").Required().String()
	nsqNumPublishers    = kingpin.Flag("nsq-num-publishers", "NSQ publishers").Envar("NSQ_NUM_PUBLISHERS").Default("100").Int()
	nsqMaxInflight      = kingpin.Flag("nsq-max-inflight", "NSQ max inflight").Envar("NSQ_MAX_INFLIGHT").Default("250").Int()

	// should be greater than prometheus scrape interval (default 30s); decreased in coding challenge
	shutdownDelay = kingpin.Flag("shutdown-delay", "shutdown delay in ms").Envar("SHUTDOWN_DELAY").Default("5000").Int()
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
	cfg.MaxInFlight = *nsqMaxInflight
	ncfg := &consumer.NSQConfig{
		NumPublishers:    *nsqNumPublishers,
		Topic:            *nsqTopic,
		Channel:          *nsqChan,
		LookupdHTTPAddrs: *nsqLookupdHTTPAddrs,
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
