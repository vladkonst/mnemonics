package respond

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, map[string]string{"hello": "world"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if body["hello"] != "world" {
		t.Errorf("body[hello] = %q, want world", body["hello"])
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad_request", "missing field")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["code"] != "bad_request" {
		t.Errorf("code = %q, want bad_request", body["code"])
	}
	if body["message"] != "missing field" {
		t.Errorf("message = %q, want missing field", body["message"])
	}
}

func TestErrorFrom(t *testing.T) {
	cases := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{apperrors.ErrNotFound, http.StatusNotFound, "not_found"},
		{apperrors.ErrAlreadyExists, http.StatusConflict, "conflict"},
		{apperrors.ErrForbidden, http.StatusForbidden, "forbidden"},
		{apperrors.ErrNotTeacher, http.StatusForbidden, "forbidden"},
		{apperrors.New("not_found", "user not found", apperrors.ErrNotFound), http.StatusNotFound, "not_found"},
	}

	for _, c := range cases {
		w := httptest.NewRecorder()
		ErrorFrom(w, c.err)

		if w.Code != c.wantStatus {
			t.Errorf("ErrorFrom(%v): status = %d, want %d", c.err, w.Code, c.wantStatus)
		}
		var body map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["code"] != c.wantCode {
			t.Errorf("ErrorFrom(%v): code = %q, want %q", c.err, body["code"], c.wantCode)
		}
	}
}

func TestErrorFrom_BadRequestError(t *testing.T) {
	w := httptest.NewRecorder()
	ErrorFrom(w, apperrors.ErrInvalidInput)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestErrorFrom_UnknownError(t *testing.T) {
	w := httptest.NewRecorder()
	ErrorFrom(w, errors.New("some unexpected internal error"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
