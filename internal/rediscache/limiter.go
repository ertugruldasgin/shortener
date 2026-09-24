package rediscache

import (
	"context"
	"log"
	"time"
)

const limitPrefix = "rate:"

// Allow increments the counter for key and reports whether it is within limit.
// Redis failures are treated as allowed, so an outage never blocks traffic.
func (c *Cache) Allow(ctx context.Context, key string, limit int, window time.Duration) bool {
	k := limitPrefix + key

	count, err := c.client.Incr(ctx, k).Result()
	if err != nil {
		log.Printf("rate limit incr %q: %v", key, err)
		return true
	}

	if count == 1 {
		if err := c.client.Expire(ctx, k, window).Err(); err != nil {
			log.Printf("rate limit expire %q: %v", key, err)
		}
	}

	return count <= int64(limit)
}
