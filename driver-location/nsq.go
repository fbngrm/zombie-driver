package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	nsq "github.com/nsqio/go-nsq"
)

var (
	nsqdTCPAddrs     = []string{"localhost:4150"}
	lookupdHTTPAddrs = []string{"localhost:4161"}
	redisAddr        = "localhost:6379"
	topic            = "locations"
	channel          = "loc-chan"
	numPublihsers    = 100
	maxInflight      = 250
)

type Location struct {
	ID   string  `json:"id"`
	Lat  float32 `json:"latitude"`
	Long float32 `json:"longitude"`
}

type LocationUpdate struct {
	UpdatedAt string  `json:"updated_at"`
	Lat       float32 `json:"latitude"`
	Long      float32 `json:"longitude"`
}

type Publisher interface {
	Publish(k, v string) error
}

type PublishHandler struct {
	Publisher
}

func (ph *PublishHandler) HandleMessage(m *nsq.Message) error {
	var l Location
	// marshal instead of decode since we expect a single JSON string
	// only not a stream or additional data
	err := json.Unmarshal(m.Body, &l)
	if err != nil {
		return err
	}
	loc := LocationUpdate{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Lat:       l.Lat,
		Long:      l.Long,
	}
	b, err := json.Marshal(loc)
	if err != nil {
		return err
	}
	return ph.Publish(l.ID, string(b))
}

type RedisPublisher struct {
	c *redis.Client
}

func (r *RedisPublisher) Publish(k, v string) error {
	err := r.c.LPush(k, v).Err()
	if err != nil {
		return err
	}
	_, err = r.GetLocations(k, 10)
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisPublisher) GetLocations(id string, count int64) ([]LocationUpdate, error) {
	locs, err := r.c.LRange(id, 0, count).Result()
	if err == redis.Nil {
		fmt.Println("key does not exist")
	} else if err != nil {
		panic(err)
	}
	fmt.Println(id, locs)
	return nil, nil
}

func main() {
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = maxInflight

	handler := &PublishHandler{
		Publisher: &RedisPublisher{
			c: redis.NewClient(&redis.Options{Addr: redisAddr}),
		},
	}
	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		log.Fatal(err)
	}
	consumer.AddConcurrentHandlers(handler, numPublihsers)

	err = consumer.ConnectToNSQDs(nsqdTCPAddrs)
	if err != nil {
		log.Fatal(err)
	}
	err = consumer.ConnectToNSQLookupds(lookupdHTTPAddrs)
	if err != nil {
		log.Fatal(err)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-consumer.StopChan:
			return
		case <-termChan:
			consumer.Stop()
		}
	}
}
