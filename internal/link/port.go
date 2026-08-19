package link

import "context"

// Repository persists and retrieves links.
type Repository interface {
	Create(ctx context.Context, l *Link) error
	BySlug(ctx context.Context, slug string) (*Link, error)
}

// Generator produces slugs for new links.
type Generator interface {
	Generate() (string, error)
}
