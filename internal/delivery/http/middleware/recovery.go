package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

// Recovery recovers from panics and returns 500 with a JSON error body.
func Recovery(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error().
						Interface("panic", rec).
						Bytes("stack", debug.Stack()).
						Msg("recovered from panic")

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"code":    "internal_error",
						"message": "internal server error",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
