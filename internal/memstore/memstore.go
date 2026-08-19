// Package memstore provides an in-memory link. Repository, used before Postgres is wired in and as a double test.
package memstore

import (
	"context"
	"ertugruldasgin/shortener/internal/link"
	"sync"
	"time"
)

// Store keep links in a map.
type Store struct {
	mu     sync.RWMutex
	links  map[string]*link.Link
	nextID int64
}

func New() *Store {
	return &Store{links: make(map[string]*link.Link)}
}

// Create stores l, returning link.ErrSlugTaken if the slug already exists.
func (s *Store) Create(ctx context.Context, l *link.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[l.Slug]; exists {
		return link.ErrSlugTaken
	}

	s.nextID++
	l.ID = s.nextID
	l.CreatedAt = time.Now()
	s.links[l.Slug] = l

	return nil
}

// BySlug returns the link for slug, or link.ErrNotFound.
func (s *Store) BySlug(ctx context.Context, slug string) (*link.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	l, ok := s.links[slug]

	if !ok {
		return nil, link.ErrNotFound
	}

	return l, nil
}
