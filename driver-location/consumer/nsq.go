package consumer

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/heetch/FabianG-technical-test/driver-location/server"
	nsq "github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
)

type Publisher interface {
	Publish(timestamp int64, key, value string) error
}

type nsqHandler struct {
	Publisher
}

func (n *nsqHandler) HandleMessage(m *nsq.Message) error {
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
	return n.Publish(t.Unix(), l.ID, string(b))
}

// RedisPublisher client representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type RedisPublisher struct {
	c *redis.Client
}

func (r *RedisPublisher) Publish(timestamp int64, key, value string) error {
	member := redis.Z{
		Score:  float64(timestamp),
		Member: value,
	}
	// O(log(N)) for each item added, where N is the number of elements in the sorted set
	return r.c.ZAddNX(key, &member).Err()
}

type NSQConfig struct {
	NumPublishers    int
	Topic            string
	Channel          string
	RedisAddr        string
	LookupdHTTPAddrs []string
	NsqdTCPAddrs     []string
	Cfg              *nsq.Config
}

type NSQ struct {
	c      *nsq.Consumer
	cfg    *NSQConfig
	logger zerolog.Logger
}

func NewNSQ(cfg *NSQConfig, logger zerolog.Logger) (*NSQ, error) {
	handler := &nsqHandler{
		Publisher: &RedisPublisher{
			c: redis.NewClient(&redis.Options{Addr: cfg.RedisAddr}),
		},
	}
	consumer, err := nsq.NewConsumer(cfg.Topic, cfg.Channel, cfg.Cfg)
	if err != nil {
		return nil, err
	}
	consumer.AddConcurrentHandlers(handler, cfg.NumPublishers)
	err = consumer.ConnectToNSQDs(cfg.NsqdTCPAddrs)
	if err != nil {
		return nil, err
	}
	err = consumer.ConnectToNSQLookupds(cfg.LookupdHTTPAddrs)
	if err != nil {
		return nil, err
	}
	logger = logger.
		With().
		Interface("topic", cfg.Topic).
		Interface("channel", cfg.Channel).
		Logger()
	return &NSQ{
		c:      consumer,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (n *NSQ) Run() {
	n.logger.Info().Msg("running nsq consumer")
	<-n.c.StopChan
}

func (n *NSQ) Shutdown() {
	n.logger.Info().Msg("stopping nsq consumer topic")
	n.c.Stop()
}
