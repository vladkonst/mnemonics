package apperrors

import (
	"errors"
	"testing"
)

func TestAppError_ErrorAndUnwrap(t *testing.T) {
	sentinel := errors.New("base error")
	err := New("some_code", "detailed message", sentinel)

	if err.Error() != "detailed message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "detailed message")
	}
	if !errors.Is(err, sentinel) {
		t.Error("Unwrap() should return the sentinel error")
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(ErrNotFound) {
		t.Error("ErrNotFound should be not found")
	}
	wrapped := New("not_found", "user not found", ErrNotFound)
	if !IsNotFound(wrapped) {
		t.Error("wrapped ErrNotFound should be not found")
	}
	if IsNotFound(ErrForbidden) {
		t.Error("ErrForbidden should not be not found")
	}
}

func TestIsConflict(t *testing.T) {
	conflictErrors := []error{
		ErrAlreadyExists,
		ErrActiveSubscriptionExists,
		ErrPaymentAlreadyProcessed,
	}
	for _, e := range conflictErrors {
		if !IsConflict(e) {
			t.Errorf("IsConflict(%v) should be true", e)
		}
	}
	if IsConflict(ErrNotFound) {
		t.Error("ErrNotFound should not be conflict")
	}
}

func TestIsForbidden(t *testing.T) {
	forbiddenErrors := []error{
		ErrForbidden,
		ErrNotTeacher,
		ErrNotYourStudent,
		ErrAccessDenied,
	}
	for _, e := range forbiddenErrors {
		if !IsForbidden(e) {
			t.Errorf("IsForbidden(%v) should be true", e)
		}
	}
	if IsForbidden(ErrNotFound) {
		t.Error("ErrNotFound should not be forbidden")
	}
}

func TestIsNotFound_Wrapped(t *testing.T) {
	wrapped := New("not_found", "theme not found", ErrNotFound)
	if !IsNotFound(wrapped) {
		t.Error("wrapped ErrNotFound should satisfy IsNotFound")
	}
}
