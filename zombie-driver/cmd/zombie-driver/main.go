package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/metrics"
	"github.com/heetch/FabianG-technical-test/zombie-driver/cmd/zombie-driver/cli"
	"github.com/heetch/FabianG-technical-test/zombie-driver/server"
)

var (
	version           = "unkown"
	service           = "zombie-driver"
	httpAddr          = ":8082"
	metricsAddr       = ":9104"
	driverLocationURL = "http://localhost:8081/drivers/%s/locations?minutes=5"
)

func main() {
	logger := cli.NewLogger(service)
	httpSrv, err := server.New(httpAddr, driverLocationURL, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s service: %v\n", service, err)
		os.Exit(2)
	}
	metricsSrv := metrics.New(metricsAddr, logger)
	cli.RunServer(httpSrv, metricsSrv)
}
