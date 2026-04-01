package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestTelegramAuth_MissingHeader(t *testing.T) {
	h := TelegramAuth()(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestTelegramAuth_InvalidHeader(t *testing.T) {
	h := TelegramAuth()(okHandler())

	cases := []string{"abc", "-1", "0", "1.5"}
	for _, v := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Telegram-User-Id", v)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("value %q: status = %d, want 401", v, w.Code)
		}
	}
}

func TestTelegramAuth_Valid(t *testing.T) {
	var capturedID int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := TelegramUserID(r.Context())
		if !ok {
			t.Error("TelegramUserID not found in context")
		}
		capturedID = id
		w.WriteHeader(http.StatusOK)
	})

	h := TelegramAuth()(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Telegram-User-Id", "12345")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if capturedID != 12345 {
		t.Errorf("capturedID = %d, want 12345", capturedID)
	}
}

func TestAdminAuth_MissingHeader(t *testing.T) {
	h := AdminAuth("secret")(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAdminAuth_WrongToken(t *testing.T) {
	h := AdminAuth("secret")(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Admin-Token", "wrong")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestAdminAuth_CorrectToken(t *testing.T) {
	h := AdminAuth("secret")(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Admin-Token", "secret")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRecovery_Panic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	h := Recovery(zerolog.Nop())(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	h := Recovery(zerolog.Nop())(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
