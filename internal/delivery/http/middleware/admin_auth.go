package middleware

import (
	"encoding/json"
	"net/http"
)

// AdminAuth checks the X-Admin-Token header against the configured token.
// Returns 401 if missing, 403 if wrong.
func AdminAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerVal := r.Header.Get("X-Admin-Token")
			if headerVal == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "unauthorized",
					"message": "X-Admin-Token header is required",
				})
				return
			}

			if headerVal != token {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "forbidden",
					"message": "invalid admin token",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
