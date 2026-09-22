package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"

	"ertugruldasgin/shortener/internal/link"
)

const (
	keyPrefix = "link:"
	opTimeout = 50 * time.Millisecond
)

var (
	hits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shortener_cache_hits_total",
		Help: "Link lookups served from the cache.",
	})
	misses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shortener_cache_misses_total",
		Help: "Link lookups that fell through to the database.",
	})
	failures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shortener_cache_errors_total",
		Help: "Cache operations that failed and were treated as misses.",
	})
)

var _ link.Cache = (*Cache)(nil)

// Cache stores links in Redis as JSON.
type Cache struct {
	client *redis.Client
}

// New connects to the Redis instance at url.
func New(url string) (*Cache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}

	// A slow cache must not slow redirects; give up fast and fall back to the database.
	opts.ReadTimeout = opTimeout
	opts.WriteTimeout = opTimeout

	return &Cache{client: redis.NewClient(opts)}, nil
}

// Ping checks the connection.
func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close releases the connection pool.
func (c *Cache) Close() error {
	return c.client.Close()
}

// Get returns the cached link for slug. Any failure is reported as a miss.
func (c *Cache) Get(ctx context.Context, slug string) (*link.Link, bool) {
	data, err := c.client.Get(ctx, keyPrefix+slug).Bytes()
	if errors.Is(err, redis.Nil) {
		misses.Inc()
		return nil, false
	}
	if err != nil {
		failures.Inc()
		misses.Inc()
		log.Printf("cache get %q: %v", slug, err)
		return nil, false
	}

	var l link.Link
	if err := json.Unmarshal(data, &l); err != nil {
		failures.Inc()
		misses.Inc()
		log.Printf("cache decode %q: %v", slug, err)
		return nil, false
	}

	hits.Inc()
	return &l, true
}

// Set stores l for ttl.
func (c *Cache) Set(ctx context.Context, l *link.Link, ttl time.Duration) {
	data, err := json.Marshal(l)
	if err != nil {
		failures.Inc()
		log.Printf("cache encode %q: %v", l.Slug, err)
		return
	}

	if err := c.client.Set(ctx, keyPrefix+l.Slug, data, ttl).Err(); err != nil {
		failures.Inc()
		log.Printf("cache set %q: %v", l.Slug, err)
	}
}

// Delete evicts slug.
func (c *Cache) Delete(ctx context.Context, slug string) {
	if err := c.client.Del(ctx, keyPrefix+slug).Err(); err != nil {
		failures.Inc()
		log.Printf("cache delete %q: %v", slug, err)
	}
}
