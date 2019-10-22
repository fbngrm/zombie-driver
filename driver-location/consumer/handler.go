package consumer

import (
	"encoding/json"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/heetch/FabianG-technical-test/types"
	nsq "github.com/nsqio/go-nsq"
)

type Publisher interface {
	Publish(timestamp int64, key string, l types.LocationUpdate) error
}

// implements the nsq.Handler interface
type LocationUpdater struct {
	Publisher
}

func (h *LocationUpdater) HandleMessage(m *nsq.Message) error {
	var l types.Location
	// marshal instead of decode since we expect a single JSON string only not
	// a stream or additional data
	err := json.Unmarshal(m.Body, &l)
	if err != nil {
		return err
	}
	t := time.Unix(0, m.Timestamp)
	lu := types.LocationUpdate{
		UpdatedAt: t.Format(time.RFC3339),
		Lat:       l.Lat,
		Long:      l.Long,
	}
	// we add a circuit breaker here although we are currently working with
	// redis handler only which does not require it
	// see: https://github.com/go-redis/redis/issues/675
	return hystrix.Do("handle_nsq_msg", func() error {
		return h.Publish(t.UnixNano(), l.ID, lu)
	}, nil)
}
