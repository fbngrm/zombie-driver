package consumer

import (
	"encoding/json"
	"time"

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
	// marshal instead of decode since we expect a single JSON string
	// only not a stream or additional data
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
	return h.Publish(t.UnixNano(), l.ID, lu)
}
