package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
)

type Server struct {
	Mux    *http.ServeMux
	server *http.Server
}

func New(cfg *config) (*Server, error) {
	s := new(http.Server)
	srv := Server{
		Mux:    http.NewServeMux(),
		server: s,
	}
	for _, url := range cfg.URLs {
		h, err := newHandler(url)
		if err != nil {
			return nil, err
		}
		srv.Mux.Handle(url.Path, h)
	}
	return &srv, nil
}

func newHandler(u URL) (http.Handler, error) {
	p, err := u.protocol()
	if err != nil {
		return nil, err
	}
	switch p {
	case NSQ:
		return &nsqHandler{}, nil
	case HTTP:
		return &httpHandler{}, nil
	default:
		return nil, fmt.Errorf("no handler found for %s", p)
	}
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

type nsqHandler struct{}

func (n *nsqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

type httpHandler struct{}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
