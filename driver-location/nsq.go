package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/heetch/FabianG-technical-test/driver-location/server"
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
	updateInteval    = 5 // seconds between location updates
)

type Publisher interface {
	Publish(timestamp int64, key, value string) error
}

type PublishHandler struct {
	Publisher
}

func (ph *PublishHandler) HandleMessage(m *nsq.Message) error {
	var l server.Location
	// marshal instead of decode since we expect a single JSON string
	// only not a stream or additional data
	err := json.Unmarshal(m.Body, &l)
	if err != nil {
		return err
	}
	t := time.Now()
	loc := server.LocationUpdate{
		UpdatedAt: t.UTC().Format(time.RFC3339),
		Lat:       l.Lat,
		Long:      l.Long,
	}
	b, err := json.Marshal(loc)
	if err != nil {
		return err
	}
	return ph.Publish(t.Unix(), l.ID, string(b))
}

type RedisPublisher struct {
	c *redis.Client
}

func (r *RedisPublisher) Publish(timestamp int64, key, value string) error {
	member := redis.Z{
		Score:  float64(timestamp),
		Member: value,
	}
	// O(log(N)) for each item added, where N is the number of elements in the sorted set.
	fmt.Printf("add %+v ", member)
	err := r.c.ZAddNX(key, &member).Err()
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)
	minutes := 5
	t := time.Now()
	min := t.Add(-1 * time.Duration(minutes) * time.Minute).Unix()
	opt := redis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(t.Unix(), 10),
	}
	locations := r.c.ZRangeByScore(key, &opt)
	fmt.Println(locations)
	return nil
}

func (r *RedisPublisher) GetLocations(id string, count int64) ([]server.LocationUpdate, error) {
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
