package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/metrics"
	"github.com/heetch/FabianG-technical-test/zombie-driver/cmd/zombie-driver/cli"
	"github.com/heetch/FabianG-technical-test/zombie-driver/server"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unkown"

	service           = kingpin.Flag("service", "service name").Envar("SERVICE").Default("zombie-driver").String()
	httpAddr          = kingpin.Flag("http-addr", "address of HTTP server").Envar("HTTP_ADDR").Default(":8082").String()
	metricsAddr       = kingpin.Flag("metrics-addr", "address of metrics server").Envar("METRICS_ADDR").Default(":9104").String()
	driverLocationURL = kingpin.Flag("driver-location-url", "address of driver-location service").Envar("DRIVER_LOCATION_URL").Required().String()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	logger := cli.NewLogger(*service, version)

	httpSrv, err := server.New(*httpAddr, *driverLocationURL, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s service: %v\n", *service, err)
		os.Exit(2)
	}
	metricsSrv := metrics.New(*metricsAddr, logger)
	cli.RunServer(httpSrv, metricsSrv)
}
