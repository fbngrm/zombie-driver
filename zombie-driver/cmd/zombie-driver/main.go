package main

import (
	"fmt"
	"os"

	"github.com/afex/hystrix-go/hystrix"
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
	zombieRadius      = kingpin.Flag("zombie-radius", "radius a zombie can move").Envar("ZOMBIE_RADIUS").Required().Float()
	zombieTime        = kingpin.Flag("zombie-time", "duration for fetching driver locations in minutes").Envar("ZOMBIE_TIME").Default("5").Int()
	// shutdownDelay is sleep time before we shutdown the server and signal arrives (ms)
	shutdownDelay = kingpin.Flag("shutdown-delay", "shutdown delay").Envar("SHUTDOWN_DELAY").Default("5000").Int()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	// configure circuit-breaker
	hystrix.ConfigureCommand("driver_location", hystrix.CommandConfig{
		Timeout:               1000, // ms
		MaxConcurrentRequests: 200,
		ErrorPercentThreshold: 25,
	})

	logger := cli.NewLogger(*service, version)

	httpSrv, err := server.New(*httpAddr, *driverLocationURL, *zombieRadius, *zombieTime, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s service: %v\n", *service, err)
		os.Exit(2)
	}
	metricsSrv := metrics.New(*metricsAddr, logger)
	cli.RunServer(httpSrv, metricsSrv, *shutdownDelay)
}
