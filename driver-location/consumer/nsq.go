package consumer

import (
	nsq "github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
)

// NSQConfig represents configuration data for an nsq consumer.
type NSQConfig struct {
	NumPublishers    int
	Topic            string
	Channel          string
	LookupdHTTPAddrs []string
	NsqdTCPAddrs     []string
	Cfg              *nsq.Config
}

// NSQ consumes a nsq channel.
type NSQ struct {
	c      *nsq.Consumer
	cfg    *NSQConfig
	logger zerolog.Logger
}

// NewNSQ returns a ready to use NSQ. It returns an error if consumer
// initilization or connecting with the nsq server fails.
func NewNSQ(cfg *NSQConfig, handler nsq.Handler, logger zerolog.Logger) (*NSQ, error) {
	consumer, err := nsq.NewConsumer(cfg.Topic, cfg.Channel, cfg.Cfg)
	if err != nil {
		return nil, err
	}
	// Todo: set logger on consumer
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

// Run waits for the consumer of n to stop.
func (n *NSQ) Run() {
	n.logger.Info().Msg("running nsq consumer")
	<-n.c.StopChan
}

// Shutdown stops the consumer of n.
func (n *NSQ) Shutdown() {
	n.logger.Info().Msg("stopping nsq consumer topic")
	n.c.Stop()
}
