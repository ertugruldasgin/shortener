// Package link contains the core domain for shortened links.
package link

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Domain errors returned by services and adapters.
var (
	ErrNotFound      = errors.New("link: not found")
	ErrSlugTaken     = errors.New("link: slug already taken")
	ErrInvalidTarget = errors.New("link: invalid target URL")
	ErrInvalidSlug   = errors.New("link: invalid slug")
	ErrExpired       = errors.New("link: expired")
)

// Link is a shortened link: a slug that redirects to a target URL.
type Link struct {
	ID        int64
	Slug      string
	Target    string
	IsCustom  bool
	CreatedAt time.Time
	ExpiresAt *time.Time // nil means never expires
}

// ValidateTarget checks that raw is usable http(s) URL.
func ValidateTarget(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("%w: empty", ErrInvalidTarget)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTarget, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrInvalidTarget)
	}

	if u.Host == "" {
		return fmt.Errorf("%w: missing host", ErrInvalidTarget)
	}

	return nil
}

var reservedSlugs = map[string]bool{
	"api":     true,
	"healthz": true,
	"metrics": true,
	"static":  true,
}

const (
	minSlugLen = 2
	maxSlugLen = 32
)

const slugAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// ValidateSlug checks that slug is usable as a custom alias.
func ValidateSlug(slug string) error {
	if len(slug) < minSlugLen || len(slug) > maxSlugLen {
		return fmt.Errorf("%w: length must be between %d-%d", ErrInvalidSlug, minSlugLen, maxSlugLen)
	}

	if reservedSlugs[strings.ToLower(slug)] {
		return fmt.Errorf("%w: reserved", ErrInvalidSlug)
	}

	for _, r := range slug {
		if !strings.ContainsRune(slugAlphabet, r) {
			return fmt.Errorf("%w: invalid character %q", ErrInvalidSlug, r)
		}
	}

	return nil
}
