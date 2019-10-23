package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/heetch/FabianG-technical-test/gateway/cmd/gateway/cli"
	"github.com/heetch/FabianG-technical-test/gateway/config"
	"github.com/heetch/FabianG-technical-test/gateway/server"
	"github.com/heetch/FabianG-technical-test/metrics"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unkown"

	service     = kingpin.Flag("service", "service name").Envar("SERVICE").Default("gateway").String()
	httpAddr    = kingpin.Flag("http-addr", "address of HTTP server").Envar("HTTP_ADDR").Default(":8080").String()
	metricsAddr = kingpin.Flag("metrics-addr", "address of metrics server").Envar("METRICS_ADDR").Default(":9102").String()

	// should be greater than prometheus scrape interval (default 30s); decreased in coding challenge
	shutdownDelay = kingpin.Flag("shutdown-delay", "shutdown delay").Envar("SHUTDOWN_DELAY").Default("5000").Int()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	// configure circuit-breaker
	hystrix.ConfigureCommand("publish_nsq", hystrix.CommandConfig{
		Timeout:               1000, // ms
		MaxConcurrentRequests: 5000,
		ErrorPercentThreshold: 25,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		<-quit
		cancel()
	}()

	cfg, err := config.FromFile("./config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *service, err)
		os.Exit(2)
	}
	logger := cli.NewLogger(*service, version)
	httpSrv, err := server.New(ctx, *httpAddr, cfg, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *service, err)
		os.Exit(2)
	}

	cli.RunServer(ctx, httpSrv, metrics.New(*metricsAddr, logger), *shutdownDelay)
}
