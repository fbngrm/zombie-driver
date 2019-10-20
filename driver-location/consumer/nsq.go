package consumer

import (
	nsq "github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
)

type NSQConfig struct {
	NumPublishers    int
	Topic            string
	Channel          string
	LookupdHTTPAddrs []string
	NsqdTCPAddrs     []string
	Cfg              *nsq.Config
}

type NSQ struct {
	c      *nsq.Consumer
	cfg    *NSQConfig
	logger zerolog.Logger
}

func NewNSQ(cfg *NSQConfig, handler nsq.Handler, logger zerolog.Logger) (*NSQ, error) {
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
