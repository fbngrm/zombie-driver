package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

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
