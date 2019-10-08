package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/gateway/api/config"
	"github.com/heetch/FabianG-technical-test/gateway/api/server"
	"github.com/heetch/FabianG-technical-test/gateway/cmd/gateway/cli"
	"github.com/heetch/FabianG-technical-test/metrics"
)

func main() {
	cfg, err := config.FromFile("./config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gateway: %v\n", err)
		os.Exit(2)
	}
	log := cli.NewLogger()
	httpSrv, err := server.New(8080, cfg, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gateway: %v\n", err)
		os.Exit(2)
	}
	metricsSrv := metrics.NewMetrics(9102, log)
	cli.RunServer(httpSrv, metricsSrv)
}
