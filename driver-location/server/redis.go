package server

import (
	"strconv"

	"github.com/go-redis/redis"
)

// redisClient wraps a Redis client, representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type redisClient struct {
	c *redis.Client
}

func (r *redisClient) ZRangeByScore(key string, min, max int64) ([]string, error) {
	opt := redis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(max, 10),
	}
	return r.c.ZRangeByScore(key, &opt).Result()
}
