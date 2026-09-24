package httpapi

import (
	"context"
	"crypto/subtle"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Limiter reports whether a key is within its request budget.
// Implementations must fail open: if the backend is unavailable, allow the request.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) bool
}

// clientIP returns the caller's address, preferring the header set by
// Cloudflare when the service runs behind a tunnel.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// rateLimit rejects requests once a client exceeds limit within window.
func rateLimit(l Limiter, name string, limit int, window time.Duration, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := name + ":" + clientIP(r)

		if !l.Allow(r.Context(), key, limit, window) {
			rateLimited.WithLabelValues(name).Inc()
			w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// requireToken rejects requests without a matching bearer token.
func requireToken(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			next(w, r)
			return
		}

		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
