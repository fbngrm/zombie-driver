package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	Mux    *mux.Router
	server *http.Server
}

func New(cfg *config) (*Server, error) {
	m := mux.NewRouter()
	for _, url := range cfg.URLs {
		h, err := newHandler(url)
		if err != nil {
			return nil, err
		}
		if url.Method == "" {
			return nil, fmt.Errorf("missing method for URL: %+v", url)
		}
		m.Handle(url.Path, h).Methods(url.Method)
	}
	return &Server{
		Mux:    m,
		server: new(http.Server),
	}, nil
}

func (s *Server) Run(ctx context.Context, addr string) error {
	s.server.Handler = s.Mux

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		fmt.Println("Listening on " + l.Addr().String())
		err := s.server.Serve(l)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	return err
}
