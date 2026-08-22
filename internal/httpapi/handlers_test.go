package httpapi

import (
	"encoding/json"
	"ertugruldasgin/shortener/internal/link"
	"ertugruldasgin/shortener/internal/memstore"
	"ertugruldasgin/shortener/internal/slug"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler() *Handler {
	store := memstore.New()
	return New(link.NewService(store, slug.New()), link.NewClickRecorder(store), "test")
}

func TestShortenReturnsCreated(t *testing.T) {
	h := newTestHandler()

	body := strings.NewReader(`{"target":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusCreated)
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

	body := strings.NewReader("not json")
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestShortenWithAlias(t *testing.T) {
	h := newTestHandler()

	body := strings.NewReader(`{"target":"https://example.com","alias":"mylink"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
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
		body := strings.NewReader(`{"target":"https://example.com","alias":"taken"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/links", body)
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

	body := strings.NewReader(`{"target":"https://example.com","alias":"api"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestShortenWithExpiry(t *testing.T) {
	h := newTestHandler()

	body := strings.NewReader(`{"target":"https://example.com","expires_in":"24h"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
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

		body := strings.NewReader(`{"target":"https://example.com","expires_in":"` + v + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/links", body)
		rec := httptest.NewRecorder()

		h.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expires_in %q: got status %d, want %d", v, rec.Code, http.StatusBadRequest)
		}
	}
}
