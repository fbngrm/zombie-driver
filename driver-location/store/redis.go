package store

import (
	"encoding/json"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/heetch/FabianG-technical-test/types"
)

// RedisClient representing a pool of zero or more underlying connections.
// It's safe for concurrent use by multiple goroutines.
type RedisClient struct {
	c *redis.Client
}

func (rc *RedisClient) ZAddNX(key string, member *redis.Z) error {
	// O(log(N)) for each item added, where N is the number
	// of elements in the sorted set
	return rc.c.ZAddNX(key, member).Err()
}

func (rc *RedisClient) ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
	return rc.c.ZRangeByScore(key, opt).Result()
}

// MiniRedis abstraction for unit tests only :( I tend to prefer a little dependency
// here over too much abstraction. So this should be replaced by a mock-library.
type MiniRedis interface {
	ZAddNX(key string, member *redis.Z) error
	ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error)
}

type Redis struct {
	MiniRedis
}

func NewRedis(addr string) *Redis {
	return &Redis{
		&RedisClient{
			c: redis.NewClient(&redis.Options{Addr: addr}),
		},
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
	return r.ZAddNX(key, &member)
}

// FetchRange returns all the elements in the sorted set at key with a score
// between min and max (including elements with score equal to min or max).
// The elements are considered to be ordered from low to high scores.
func (r *Redis) FetchRange(key string, min, max int64) ([]string, error) {
	opt := redis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(max, 10),
	}
	return r.ZRangeByScore(key, &opt)
}
