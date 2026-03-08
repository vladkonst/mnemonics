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
