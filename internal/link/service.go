package link

import (
	"context"
	"errors"
	"fmt"
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
