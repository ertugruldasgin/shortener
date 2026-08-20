package httpapi

import (
	"ertugruldasgin/shortener/internal/link"
	"ertugruldasgin/shortener/internal/memstore"
	"ertugruldasgin/shortener/internal/slug"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler() *Handler {
	return New(link.NewService(memstore.New(), slug.New()), "test")
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
