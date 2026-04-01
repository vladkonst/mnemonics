package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type contextKey string

const telegramUserIDKey contextKey = "telegram_user_id"

// TelegramAuth extracts X-Telegram-User-Id header, parses it as int64, and stores it in context.
// Returns 401 if the header is missing or invalid.
func TelegramAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerVal := r.Header.Get("X-Telegram-User-Id")
			if headerVal == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "unauthorized",
					"message": "X-Telegram-User-Id header is required",
				})
				return
			}

			userID, err := strconv.ParseInt(headerVal, 10, 64)
			if err != nil || userID <= 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "unauthorized",
					"message": "X-Telegram-User-Id must be a valid positive integer",
				})
				return
			}

			ctx := context.WithValue(r.Context(), telegramUserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TelegramUserID retrieves the Telegram user ID stored in the context by TelegramAuth.
func TelegramUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(telegramUserIDKey).(int64)
	return id, ok
}

// MaxBody limits the request body to 1 MB to prevent resource exhaustion.
func MaxBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}

// RequireOwner checks that the authenticated user matches pathUserID.
// Writes 403 Forbidden and returns false if they differ.
func RequireOwner(w http.ResponseWriter, r *http.Request, pathUserID int64) bool {
	authID, ok := TelegramUserID(r.Context())
	if !ok || authID != pathUserID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":    "forbidden",
			"message": "access to another user's resource is not allowed",
		})
		return false
	}
	return true
}
