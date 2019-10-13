package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/gateway/api/config"
	"github.com/heetch/FabianG-technical-test/gateway/api/server"
	"github.com/heetch/FabianG-technical-test/gateway/cmd/gateway/cli"
	"github.com/heetch/FabianG-technical-test/metrics"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unkown"

	service     = kingpin.Flag("service", "service name").Envar("SERVICE").Default("gateway").String()
	httpAddr    = kingpin.Flag("http-addr", "address of HTTP server").Envar("HTTP_ADDR").Default(":8080").String()
	metricsAddr = kingpin.Flag("metrics-addr", "address of metrics server").Envar("METRICS_ADDR").Default(":9102").String()

	// shutdownDelay is sleep time before we shutdown the server and signal arrives (ms)
	shutdownDelay = kingpin.Flag("shutdown-delay", "shutdown delay").Envar("SHUTDOWN_DELAY").Default("5000").Int()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	cfg, err := config.FromFile("./config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *service, err)
		os.Exit(2)
	}
	logger := cli.NewLogger(*service, version)
	httpSrv, err := server.New(*httpAddr, cfg, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *service, err)
		os.Exit(2)
	}
	cli.RunServer(httpSrv, metrics.New(*metricsAddr, logger), *shutdownDelay)
}
