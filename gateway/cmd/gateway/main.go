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
	httpAddr    = ":8081"
	metricsAddr = ":9103"
)

func main() {
	cfg, err := config.FromFile("./config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", service, err)
		os.Exit(2)
	}
	log := cli.NewLogger()
	httpSrv, err := server.New(httpAddr, cfg, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", service, err)
		os.Exit(2)
	}
	metricsSrv := metrics.NewMetrics(metricsAddr, log)
	cli.RunServer(httpSrv, metricsSrv)
}
