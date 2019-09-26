package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	nsq "github.com/nsqio/go-nsq"
)

// TODO: CORS
func newHandler(u URL) (http.Handler, error) {
	p, err := u.protocol()
	if err != nil {
		return nil, err
	}
	switch p {
	case NSQ:
		return newNSQHandler(u)
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

type nsqHandler struct {
	topic     string
	producers map[string]*nsq.Producer
}

func newNSQHandler(u URL) (*nsqHandler, error) {
	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("go-nsq/%s", nsq.VERSION)
	producers := make(map[string]*nsq.Producer)
	for _, addr := range u.NSQ.TCPAddrs {
		producer, err := nsq.NewProducer(addr, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create nsq.Producer - %s", err)
		}
		producers[addr] = producer
	}
	return &nsqHandler{
		topic:     u.NSQ.Topic,
		producers: producers,
	}, nil
}

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
