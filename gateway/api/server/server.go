package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// TODO: CORS
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

type location struct {
	Lat  float64 `json:"latitude"`
	Long float64 `json:"longitude"`
}

type nsqHandler struct{}

func (n *nsqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// FIXME
		panic(err)
	}
	var l location
	err = json.Unmarshal(body, &l)
	if err != nil {
		// FIXME
		panic(err)
	}
	vars := mux.Vars(r)
	fmt.Fprintf(w, "Hi %s, I am at %+v!", vars["id"], l)
}

type httpHandler struct{}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
