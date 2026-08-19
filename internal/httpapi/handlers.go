// Package httpapi exposes the link service over HTTP.
package httpapi

import (
	"encoding/json"
	"errors"
	"ertugruldasgin/shortener/internal/link"
	"net/http"
	"time"
)

// Handler serves the link HTTP API.
type Handler struct {
	svc *link.Service
}

func New(svc *link.Service) *Handler {
	return &Handler{svc: svc}
}

// Routes returns the router with all endpoints registered.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/links", h.shorten)
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /slug", h.redirect)

	return mux
}

type shortenRequest struct {
	Target string `json:"target"`
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

	l, err := h.svc.Shorten(r.Context(), req.Target)
	if err != nil {
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
	w.WriteHeader(http.StatusOK)
}
