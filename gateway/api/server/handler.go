package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

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
		// in a real world scenario we would factor this out to perform more
		// sofisticated operations like reqriting headers for https requests etc.
		return httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   u.HTTP.Host,
		}), nil
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
	// marshal instead of decode since we expect a single JSON string
	// only not a stream or additional data
	err = json.Unmarshal(body, &l)
	if err != nil {
		// FIXME
		panic(err)
	}
	// relies on sane input for 'id'; currently sanitized by mux only
	l.ID = mux.Vars(r)["id"]
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
