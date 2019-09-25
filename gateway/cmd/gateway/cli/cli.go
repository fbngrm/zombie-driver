package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/heetch/FabianG-technical-test/gateway/api/server"
)

func RunServer(srv *server.Server, addr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		<-quit
		cancel()
	}()

	return srv.Run(ctx, addr)
}
