package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// RequestID adds an X-Request-Id header (UUID) to every request and response.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = uuid.NewString()
			}
			r.Header.Set("X-Request-Id", id)
			w.Header().Set("X-Request-Id", id)
			next.ServeHTTP(w, r)
		})
	}
}
