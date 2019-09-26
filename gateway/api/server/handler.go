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

// TODO: evaluate value range for fields
type location struct {
	ID   string  `json:"id"`
	Lat  float32 `json:"latitude"`
	Long float32 `json:"longitude"`
}

type nsqHandler struct {
	topic     string
	producers map[string]*nsq.Producer
}

// NOTE: throttling is not enabled for producers which should be done in a
// production environment.
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
	l.ID = mux.Vars(r)["id"] // relies on sane input for 'id'
	b, err := json.Marshal(l)
	if err != nil {
		// FIXME
		panic(err)
	}
	for _, producer := range n.producers {
		err := producer.Publish(n.topic, b)
		if err != nil {
			// FIXME
			panic(err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

type httpHandler struct{}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
