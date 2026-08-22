package memstore

import (
	"context"
	"errors"
	"ertugruldasgin/shortener/internal/link"
	"testing"
)

func TestCreateAndBySlug(t *testing.T) {
	s := New()
	ctx := context.Background()

	slug := "abc123"
	target := "https://example.com"

	l := &link.Link{Slug: slug, Target: target}
	if err := s.Create(ctx, l); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	got, err := s.BySlug(ctx, slug)
	if err != nil {
		t.Fatalf("BySlug(): %v", err)
	}

	if got.Target != target {
		t.Errorf("got target %q, want %q", got.Target, target)
	}
}

func TestCreateDuplicateSlug(t *testing.T) {
	s := New()
	ctx := context.Background()

	slug := "dup"

	_ = s.Create(ctx, &link.Link{Slug: slug, Target: "https://a.com"})
	err := s.Create(ctx, &link.Link{Slug: slug, Target: "https://b.com"})

	if !errors.Is(err, link.ErrSlugTaken) {
		t.Errorf("got %v, want ErrSlugTaken", err)
	}
}

func TestBySlugNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.BySlug(ctx, "missing")
	if !errors.Is(err, link.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestRecordClick(t *testing.T) {
	s := New()
	ctx := context.Background()

	l := &link.Link{Slug: "abc123", Target: "https://example.com"}
	if err := s.Create(ctx, l); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	c := &link.Click{LinkID: l.ID, Referrer: "https://news.example"}
	if err := s.RecordClick(ctx, c); err != nil {
		t.Errorf("RecordClick(): %v", err)
	}
}
