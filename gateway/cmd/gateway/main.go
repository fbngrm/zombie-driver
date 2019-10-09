package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/gateway/api/config"
	"github.com/heetch/FabianG-technical-test/gateway/api/server"
	"github.com/heetch/FabianG-technical-test/gateway/cmd/gateway/cli"
	"github.com/heetch/FabianG-technical-test/metrics"
)

var (
	service     = "gateway"
	httpAddr    = ":8080"
	metricsAddr = ":9102"
)

func main() {
	cfg, err := config.FromFile("./config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", service, err)
		os.Exit(2)
	}
	logger := cli.NewLogger(service)
	httpSrv, err := server.New(httpAddr, cfg, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", service, err)
		os.Exit(2)
	}
	metricsSrv := metrics.New(metricsAddr, logger)
	cli.RunServer(httpSrv, metricsSrv)
}
