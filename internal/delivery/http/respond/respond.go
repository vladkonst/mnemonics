// Package respond provides HTTP response helpers for the delivery layer.
package respond

import (
	"encoding/json"
	"net/http"

	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// JSON writes status and data as JSON to the response writer.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Error writes a structured JSON error response.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, map[string]string{
		"code":    code,
		"message": message,
	})
}

// ErrorFrom maps an apperror to the appropriate HTTP status code and writes the response.
func ErrorFrom(w http.ResponseWriter, err error) {
	switch {
	case apperrors.IsNotFound(err):
		Error(w, http.StatusNotFound, "not_found", err.Error())
	case apperrors.IsConflict(err):
		Error(w, http.StatusConflict, "conflict", err.Error())
	case apperrors.IsForbidden(err):
		Error(w, http.StatusForbidden, "forbidden", err.Error())
	case apperrors.IsBadRequest(err):
		Error(w, http.StatusBadRequest, "bad_request", err.Error())
	default:
		Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
