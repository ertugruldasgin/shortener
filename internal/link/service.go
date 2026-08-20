package link

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const maxSlugAttempts = 5

// Service implements the link use cases.
type Service struct {
	repo Repository
	gen  Generator
}

func NewService(repo Repository, gen Generator) *Service {
	return &Service{repo: repo, gen: gen}
}

// Shorten creates a new link for target, retrying on slug collisions.
func (s *Service) Shorten(ctx context.Context, target string) (*Link, error) {
	target = strings.TrimSpace(target)
	if err := ValidateTarget(target); err != nil {
		return nil, err
	}

	for range maxSlugAttempts {
		slug, err := s.gen.Generate()
		if err != nil {
			return nil, fmt.Errorf("generating slug: %w", err)
		}

		l := &Link{Slug: slug, Target: target}

		err = s.repo.Create(ctx, l)
		if err == nil {
			return l, nil
		}

		if !errors.Is(err, ErrSlugTaken) {
			return nil, fmt.Errorf("creating link: %w", err)
		}
	}

	return nil, fmt.Errorf("no free slug after %d attempts", maxSlugAttempts)
}

// Resolve returns the link for slug, or ErrExpired if it is no longer valid.
func (s *Service) Resolve(ctx context.Context, slug string, now time.Time) (*Link, error) {
	l, err := s.repo.BySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if l.ExpiresAt != nil && !now.Before(*l.ExpiresAt) {
		return nil, ErrExpired
	}

	return l, nil
}
