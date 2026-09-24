package httpapi

import (
	"context"
	"encoding/json"
	"ertugruldasgin/shortener/internal/link"
	"ertugruldasgin/shortener/internal/memstore"
	"ertugruldasgin/shortener/internal/slug"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	clickBufferSize = 256
	testToken       = "test-token"
	testAdminToken  = "admin-token"
)

type allowAll struct{}

func (allowAll) Allow(context.Context, string, int, time.Duration) bool { return true }

func newTestHandler() *Handler {
	store := memstore.New()
	return New(Config{
		Service:    link.NewService(store, slug.New(), nil),
		Recorder:   link.NewClickRecorder(store, clickBufferSize),
		Limiter:    allowAll{},
		RateLimits: RateLimits{Create: 1000, Redirect: 1000, Window: time.Minute},
		Version:    "test",
		APIToken:   testToken,
		AdminToken: testAdminToken,
	})
}

// newAuthedPost builds a POST request to /api/links with a valid token.
func newAuthedPost(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+testToken)
	return req
}

func TestShortenReturnsCreated(t *testing.T) {
	h := newTestHandler()

	req := newAuthedPost(`{"target":"https://example.com"}`)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestShortenRequiresToken(t *testing.T) {
	h := newTestHandler()

	body := strings.NewReader(`{"target":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestShortenRejectsWrongToken(t *testing.T) {
	h := newTestHandler()

	body := strings.NewReader(`{"target":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRedirectNotFound(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestShortenInvalidJSON(t *testing.T) {
	h := newTestHandler()

	req := newAuthedPost("not json")
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestShortenWithAlias(t *testing.T) {
	h := newTestHandler()

	req := newAuthedPost(`{"target":"https://example.com","alias":"mylink"}`)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp shortenResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Slug != "mylink" {
		t.Errorf("got slug %q, want %q", resp.Slug, "mylink")
	}
}

func TestShortenAliasConflict(t *testing.T) {
	h := newTestHandler()

	post := func() int {
		req := newAuthedPost(`{"target":"https://example.com","alias":"taken"}`)
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, req)
		return rec.Code
	}

	post()
	if got := post(); got != http.StatusConflict {
		t.Errorf("got status %d, want %d", got, http.StatusConflict)
	}
}

func TestShortenReservedAlias(t *testing.T) {
	h := newTestHandler()

	req := newAuthedPost(`{"target":"https://example.com","alias":"api"}`)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestShortenWithExpiry(t *testing.T) {
	h := newTestHandler()

	req := newAuthedPost(`{"target":"https://example.com","expires_in":"24h"}`)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp shortenResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.ExpiresAt == nil {
		t.Error("expected expires_at in response, got nil")
	}
}

func TestShortenInvalidExpiry(t *testing.T) {
	for _, v := range []string{"soon", "-1h", "0s"} {
		h := newTestHandler()

		req := newAuthedPost(`{"target":"https://example.com","expires_in":"` + v + `"}`)
		rec := httptest.NewRecorder()

		h.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expires_in %q: got status %d, want %d", v, rec.Code, http.StatusBadRequest)
		}
	}
}

// denyAll rejects every request.
type denyAll struct{}

func (denyAll) Allow(context.Context, string, int, time.Duration) bool { return false }

func TestRedirectRateLimited(t *testing.T) {
	store := memstore.New()
	h := New(Config{
		Service:    link.NewService(store, slug.New(), nil),
		Recorder:   link.NewClickRecorder(store, clickBufferSize),
		Limiter:    denyAll{},
		RateLimits: RateLimits{Create: 1, Redirect: 1, Window: time.Minute},
		Version:    "test",
		APIToken:   testToken,
		AdminToken: testAdminToken,
	})

	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}
