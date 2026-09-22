package link

import (
	"context"
	"time"
)

// Repository persists and retrieves links.
type Repository interface {
	Create(ctx context.Context, l *Link) error
	BySlug(ctx context.Context, slug string) (*Link, error)
	Delete(ctx context.Context, slug string) error
	RecordClick(ctx context.Context, c *Click) error
	RecordClicks(ctx context.Context, cs []Click) error
}

// Cache holds recently resolved links in front of the repository.
// Implementations must treat failures as misses: a broken cache degrades to database reads, it never fails a request.
type Cache interface {
	Get(ctx context.Context, slug string) (*Link, bool)
	Set(ctx context.Context, l *Link, ttl time.Duration)
	Delete(ctx context.Context, slug string)
}

// Generator produces slugs for new links.
type Generator interface {
	Generate() (string, error)
}

// NopCache is a Cache that stores nothing.
type NopCache struct{}

func (NopCache) Get(context.Context, string) (*Link, bool) { return nil, false }
func (NopCache) Set(context.Context, *Link, time.Duration) {}
func (NopCache) Delete(context.Context, string)            {}
