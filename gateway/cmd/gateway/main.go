package main

import (
	"fmt"
	"os"

	"github.com/heetch/FabianG-technical-test/gateway/api/server"
	"github.com/heetch/FabianG-technical-test/gateway/cmd/gateway/cli"
)

func main() {
	cfg, err := server.LoadConfig("./config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gateway: %v\n", err)
		os.Exit(2)
	}
	srv := server.New(cfg)
	cli.RunServer(srv, "127.0.0.1:8080")
}
