package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

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
