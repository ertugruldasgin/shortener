package link

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeGen returns the given slugs in order.
type fakeGen struct {
	slugs []string
	calls int
}

func (f *fakeGen) Generate() (string, error) {
	s := f.slugs[f.calls]
	f.calls++
	return s, nil
}

type fakeRepo struct {
	failFirst int
	calls     int
	saved     *Link
	stored    map[string]*Link
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{stored: make(map[string]*Link)}
}

func (f *fakeRepo) Create(ctx context.Context, l *Link) error {
	f.calls++
	if f.calls <= f.failFirst {
		return ErrSlugTaken
	}

	f.saved = l
	f.stored[l.Slug] = l
	return nil
}

func (f *fakeRepo) BySlug(ctx context.Context, slug string) (*Link, error) {
	l, ok := f.stored[slug]
	if !ok {
		return nil, ErrNotFound
	}
	return l, nil
}

func (f *fakeRepo) RecordClick(ctx context.Context, c *Click) error {
	return nil
}

func TestCreateGeneratesSlug(t *testing.T) {
	gen := &fakeGen{slugs: []string{"abc123"}}
	repo := newFakeRepo()
	svc := NewService(repo, gen)

	got, err := svc.Create(context.Background(), CreateRequest{Target: "https://example.com"})
	if err != nil {
		t.Fatalf("Create(): %v", err)
	}

	if got.Slug != "abc123" {
		t.Errorf("got slug %q, want %q", got.Slug, "abc123")
	}
	if got.IsCustom {
		t.Error("got IsCustom true, want false")
	}
}

func TestCreateRetriesOnCollision(t *testing.T) {
	gen := &fakeGen{slugs: []string{"aaa111", "bbb222", "ccc333"}}
	repo := newFakeRepo()
	repo.failFirst = 2
	svc := NewService(repo, gen)

	got, err := svc.Create(context.Background(), CreateRequest{Target: "https://example.com"})
	if err != nil {
		t.Fatalf("Create(): %v", err)
	}

	if got.Slug != "ccc333" {
		t.Errorf("got slug %q, want %q", got.Slug, "ccc333")
	}
	if repo.calls != 3 {
		t.Errorf("repo called %d times, want 3", repo.calls)
	}
}

func TestCreateGivesUpAfterMaxAttempts(t *testing.T) {
	gen := &fakeGen{slugs: []string{"a", "b", "c", "d", "e"}}
	repo := newFakeRepo()
	repo.failFirst = 99
	svc := NewService(repo, gen)

	_, err := svc.Create(context.Background(), CreateRequest{Target: "https://example.com"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if repo.calls != maxSlugAttempts {
		t.Errorf("repo called %d times, want %d", repo.calls, maxSlugAttempts)
	}
}

func TestCreateWithCustomSlugDoesNotRetry(t *testing.T) {
	gen := &fakeGen{}
	repo := newFakeRepo()
	repo.failFirst = 1
	svc := NewService(repo, gen)

	_, err := svc.Create(context.Background(), CreateRequest{
		Target: "https://example.com",
		Slug:   "mine",
	})

	if !errors.Is(err, ErrSlugTaken) {
		t.Errorf("got %v, want ErrSlugTaken", err)
	}
	if repo.calls != 1 {
		t.Errorf("repo called %d times, want 1", repo.calls)
	}
}

func TestCreateRejectsInvalidTarget(t *testing.T) {
	svc := NewService(newFakeRepo(), &fakeGen{})

	_, err := svc.Create(context.Background(), CreateRequest{Target: "merhaba"})
	if !errors.Is(err, ErrInvalidTarget) {
		t.Errorf("got %v, want ErrInvalidTarget", err)
	}
}

func TestResolveExpired(t *testing.T) {
	gen := &fakeGen{slugs: []string{"exp123"}}
	repo := newFakeRepo()
	svc := NewService(repo, gen)

	past := time.Now().Add(-time.Hour)
	_, err := svc.Create(context.Background(), CreateRequest{
		Target:    "https://example.com",
		ExpiresAt: &past,
	})
	if err != nil {
		t.Fatalf("Create(): %v", err)
	}

	_, err = svc.Resolve(context.Background(), "exp123", time.Now())
	if !errors.Is(err, ErrExpired) {
		t.Errorf("got %v, want ErrExpired", err)
	}
}

func TestResolveNotExpired(t *testing.T) {
	gen := &fakeGen{slugs: []string{"live12"}}
	repo := newFakeRepo()
	svc := NewService(repo, gen)

	future := time.Now().Add(time.Hour)
	_, err := svc.Create(context.Background(), CreateRequest{
		Target:    "https://example.com",
		ExpiresAt: &future,
	})
	if err != nil {
		t.Fatalf("Create(): %v", err)
	}

	if _, err := svc.Resolve(context.Background(), "live12", time.Now()); err != nil {
		t.Errorf("Resolve(): unexpected error %v", err)
	}
}
