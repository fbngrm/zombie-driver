package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/driver-location/cmd/cli"
	"github.com/heetch/FabianG-technical-test/driver-location/server"
	"github.com/heetch/FabianG-technical-test/metrics"
)

func main() {
	logger := cli.NewLogger()
	httpSrv, err := server.New(8081, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "driver-location: %v\n", err)
		os.Exit(2)
	}
	metricsSrv := metrics.NewMetrics(9103, logger)
	cli.RunServer(httpSrv, metricsSrv)
}
