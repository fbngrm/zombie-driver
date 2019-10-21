package store

import (
	"encoding/json"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/heetch/FabianG-technical-test/types"
)

// Redis client representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type Redis struct {
	c *redis.Client
}

func NewRedis(addr string) *Redis {
	return &Redis{
		c: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (r *Redis) Publish(timestamp int64, key string, l types.LocationUpdate) error {
	value, err := json.Marshal(l)
	if err != nil {
		return err
	}
	member := redis.Z{
		Score:  float64(timestamp),
		Member: string(value),
	}
	// O(log(N)) for each item added, where N is the number
	// of elements in the sorted set
	return r.c.ZAddNX(key, &member).Err()
}

// FetchRange returns all the elements in the sorted set at key with a score
// between min and max (including elements with score equal to min or max).
// The elements are considered to be ordered from low to high scores.
func (r *Redis) FetchRange(key string, min, max int64) ([]string, error) {
	opt := redis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(max, 10),
	}
	return r.c.ZRangeByScore(key, &opt).Result()
}
