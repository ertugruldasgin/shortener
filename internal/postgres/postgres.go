package postgres

import (
	"context"
	"errors"
	"ertugruldasgin/shortener/internal/link"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uniqueViolation is the PostgreSQL error code for a unique constraint breach.
const uniqueViolation = "23505"

// Repo stores links in Postgre.
type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Create inserts l, translating a slug conflict into link.ErrSlugTaken.
func (r *Repo) Create(ctx context.Context, l *link.Link) error {
	const q = `
    INSERT INTO links (slug, target, is_custom, expires_at)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, q, l.Slug, l.Target, l.IsCustom, l.ExpiresAt).
		Scan(&l.ID, &l.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return link.ErrSlugTaken
	}

	if err != nil {
		return fmt.Errorf("inserting link: %w", err)
	}

	return nil
}

// BySlug returns the link for slug, or link.ErrNotFound.
func (r *Repo) BySlug(ctx context.Context, slug string) (*link.Link, error) {
	const q = `
		SELECT id, slug, target, is_custom, created_at, expires_at
		FROM links
		WHERE slug = $1`

	var l link.Link

	err := r.pool.QueryRow(ctx, q, slug).
		Scan(&l.ID, &l.Slug, &l.Target, &l.IsCustom, &l.CreatedAt, &l.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, link.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("selecting link: %w", err)
	}

	return &l, nil
}
