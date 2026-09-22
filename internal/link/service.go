package link

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const maxSlugAttempts = 5
const cacheTTL = time.Hour

// Service implements the link use cases.
type Service struct {
	repo  Repository
	gen   Generator
	cache Cache
}

func NewService(repo Repository, gen Generator, cache Cache) *Service {
	if cache == nil {
		cache = NopCache{}
	}
	return &Service{repo: repo, gen: gen, cache: cache}
}

type CreateRequest struct {
	Target    string
	Slug      string     // empty means a random slug is generated
	ExpiresAt *time.Time // nil means link never expires
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Link, error) {
	target := strings.TrimSpace(req.Target)
	if err := ValidateTarget(target); err != nil {
		return nil, err
	}

	if slug := strings.TrimSpace(req.Slug); slug != "" {
		if err := ValidateSlug(slug); err != nil {
			return nil, err
		}

		l := &Link{Target: target, Slug: slug, IsCustom: true, ExpiresAt: req.ExpiresAt}
		if err := s.repo.Create(ctx, l); err != nil {
			return nil, err
		}

		return l, nil
	}

	for range maxSlugAttempts {
		slug, err := s.gen.Generate()
		if err != nil {
			return nil, fmt.Errorf("generating slug: %w", err)
		}

		l := &Link{Slug: slug, Target: target, ExpiresAt: req.ExpiresAt}

		err = s.repo.Create(ctx, l)
		if err == nil {
			return l, nil
		}

		if !errors.Is(err, ErrSlugTaken) {
			return nil, fmt.Errorf("creating link: %w", err)
		}
	}

	return nil, fmt.Errorf("no free slugs after %d attempts", maxSlugAttempts)
}

// Delete removes the link for slug.
func (s *Service) Delete(ctx context.Context, slug string) error {
	slug = strings.TrimSpace(slug)
	if err := s.repo.Delete(ctx, slug); err != nil {
		return err
	}

	s.cache.Delete(ctx, slug)
	return nil
}

// Resolve returns the link for slug, or ErrExpired if it is no longer valid.
func (s *Service) Resolve(ctx context.Context, slug string, now time.Time) (*Link, error) {
	l, ok := s.cache.Get(ctx, slug)
	if !ok {
		var err error
		l, err = s.repo.BySlug(ctx, slug)
		if err != nil {
			return nil, err
		}

		if ttl := cacheTTLFor(l, now); ttl > 0 {
			s.cache.Set(ctx, l, ttl)
		}
	}

	if l.ExpiresAt != nil && !now.Before(*l.ExpiresAt) {
		return nil, ErrExpired
	}

	return l, nil
}

// cacheTTLFor caps the cache lifetime at the link's own expiry, so the cache
// never serves a link past its deadline.
func cacheTTLFor(l *Link, now time.Time) time.Duration {
	ttl := cacheTTL
	if l.ExpiresAt != nil {
		if remaining := l.ExpiresAt.Sub(now); remaining < ttl {
			ttl = remaining
		}
	}
	return ttl
}
