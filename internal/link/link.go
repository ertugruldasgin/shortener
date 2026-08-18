// Package link contains the core domain for shortened links.
package link

import (
	"errors"
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
