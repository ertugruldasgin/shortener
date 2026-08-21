// Package httpapi exposes the link service over HTTP.
package httpapi

import (
	"encoding/json"
	"errors"
	"ertugruldasgin/shortener/internal/link"
	"log"
	"net/http"
	"time"
)

// Handler serves the link HTTP API.
type Handler struct {
	svc     *link.Service
	version string
}

func New(svc *link.Service, version string) *Handler {
	return &Handler{svc: svc, version: version}
}

// Routes returns the router with all endpoints registered.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/links", h.shorten)
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /{slug}", h.redirect)

	return mux
}

type shortenRequest struct {
	Target string `json:"target"`
	Alias  string `json:"alias,omitempty"`
}

type shortenResponse struct {
	Slug   string `json:"slug"`
	Target string `json:"target"`
}

// shorten creates a link for the target URL in the request body.
func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	var (
		l   *link.Link
		err error
	)

	if req.Alias != "" {
		l, err = h.svc.ShortenWithSlug(r.Context(), req.Target, req.Alias)
	} else {
		l, err = h.svc.Shorten(r.Context(), req.Target)
	}

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
	json.NewEncoder(w).Encode(shortenResponse{Slug: l.Slug, Target: l.Target})

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
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, l.Target, http.StatusTemporaryRedirect)
}

// health reports that the server is up.
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": h.version})
}
