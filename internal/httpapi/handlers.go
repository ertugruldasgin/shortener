// Package httpapi exposes the link service over HTTP.
package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ertugruldasgin/shortener/internal/link"
)

// Handler serves the link HTTP API.
type Handler struct {
	svc        *link.Service
	recorder   *link.ClickRecorder
	limiter    Limiter
	limits     RateLimits
	version    string
	apiToken   string
	adminToken string
}

// RateLimits configures the per-client request budgets.
type RateLimits struct {
	Create   int
	Redirect int
	Window   time.Duration
}

// Config holds the handler's dependencies and settings.
type Config struct {
	Service    *link.Service
	Recorder   *link.ClickRecorder
	Limiter    Limiter
	RateLimits RateLimits
	Version    string
	APIToken   string
	AdminToken string
}

func New(cfg Config) *Handler {
	return &Handler{
		svc:        cfg.Service,
		recorder:   cfg.Recorder,
		limiter:    cfg.Limiter,
		limits:     cfg.RateLimits,
		version:    cfg.Version,
		apiToken:   cfg.APIToken,
		adminToken: cfg.AdminToken,
	}
}

// Routes returns the router with all endpoints registered.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/links", withMetrics("shorten",
		rateLimit(h.limiter, "shorten", h.limits.Create, h.limits.Window,
			requireToken(h.apiToken, h.shorten))))

	mux.HandleFunc("DELETE /api/links/{slug}", withMetrics("delete",
		requireToken(h.adminToken, h.deleteLink)))

	mux.HandleFunc("GET /healthz", withMetrics("health", h.health))
	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("GET /{slug}", withMetrics("redirect",
		rateLimit(h.limiter, "redirect", h.limits.Redirect, h.limits.Window,
			h.redirect)))

	return mux
}

type shortenRequest struct {
	Target    string `json:"target"`
	Alias     string `json:"alias,omitempty"`
	ExpiresIn string `json:"expires_in,omitempty"`
}

type shortenResponse struct {
	Slug      string     `json:"slug"`
	Target    string     `json:"target"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// shorten creates a link for the target URL in the request body.
func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err != nil || d <= 0 {
			http.Error(w, "invalid expires_in", http.StatusBadRequest)
			return
		}
		t := time.Now().Add(d)
		expiresAt = &t
	}

	l, err := h.svc.Create(r.Context(), link.CreateRequest{
		Target:    req.Target,
		Slug:      req.Alias,
		ExpiresAt: expiresAt,
	})

	switch {
	case errors.Is(err, link.ErrInvalidTarget):
		http.Error(w, "invalid target URL", http.StatusBadRequest)
		return
	case errors.Is(err, link.ErrInvalidSlug):
		http.Error(w, "invalid alias", http.StatusBadRequest)
		return
	case errors.Is(err, link.ErrSlugTaken):
		http.Error(w, "alias already taken", http.StatusConflict)
		return
	case err != nil:
		log.Printf("shorten: %v", err)
		http.Error(w, "could not create link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shortenResponse{
		Slug:      l.Slug,
		Target:    l.Target,
		ExpiresAt: l.ExpiresAt,
	})
}

// redirect sends the client to the target of slug.
func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	l, err := h.svc.Resolve(r.Context(), slug, time.Now())
	switch {
	case errors.Is(err, link.ErrNotFound):
		http.NotFound(w, r)
		return
	case errors.Is(err, link.ErrExpired):
		http.Error(w, "link expired", http.StatusGone)
		return
	case err != nil:
		log.Printf("redirect: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.recorder.Record(link.Click{
		LinkID:    l.ID,
		Referrer:  r.Referer(),
		UserAgent: r.UserAgent(),
	})
	clicksAttempted.Inc()

	http.Redirect(w, r, l.Target, http.StatusTemporaryRedirect)
}

// deleteLink removes the link for the slug in the path.
func (h *Handler) deleteLink(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	err := h.svc.Delete(r.Context(), slug)
	switch {
	case errors.Is(err, link.ErrNotFound):
		http.NotFound(w, r)
		return
	case err != nil:
		log.Printf("delete: %v", err)
		http.Error(w, "could not delete link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// health reports that the server is up.
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": h.version})
}
