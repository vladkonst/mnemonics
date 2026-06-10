package apperrors

import "errors"

// Sentinel domain errors used across all layers.
// Delivery layer maps these to HTTP status codes.
var (
	// 404
	ErrNotFound = errors.New("not found")
	// 409
	ErrAlreadyExists   = errors.New("already exists")
	ErrAlreadyConsumed = errors.New("already consumed")
	// 403
	ErrForbidden       = errors.New("forbidden")
	ErrNotTeacher      = errors.New("user is not a teacher")
	ErrNotYourStudent  = errors.New("student does not belong to this teacher")
	// 400
	ErrInvalidInput        = errors.New("invalid input")
	ErrInviteLinkExhausted = errors.New("invite link has no remaining activations")
	ErrInvalidScore        = errors.New("score must be between 0 and 100")
	ErrInvalidStatus      = errors.New("invalid status transition")
	ErrAccessDenied       = errors.New("theme access denied: complete previous theme first")
	// 409 payment
	ErrActiveSubscriptionExists = errors.New("user already has an active subscription")
	ErrPaymentAlreadyProcessed  = errors.New("payment already processed")
	// 409 corporate groups
	ErrGroupAlreadyClaimed = errors.New("teacher slot already claimed for this group")
)

// AppError wraps a sentinel error with an HTTP-friendly code and message.
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// New creates an AppError from a sentinel.
func New(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// IsNotFound reports whether err is (or wraps) ErrNotFound.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsConflict reports whether err is a conflict-type error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrAlreadyExists) ||
		errors.Is(err, ErrAlreadyConsumed) ||
		errors.Is(err, ErrActiveSubscriptionExists) ||
		errors.Is(err, ErrPaymentAlreadyProcessed) ||
		errors.Is(err, ErrGroupAlreadyClaimed)
}

// IsBadRequest reports whether err is a client-error (400) type error.
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInviteLinkExhausted) ||
		errors.Is(err, ErrInvalidScore) ||
		errors.Is(err, ErrInvalidStatus)
}

// IsForbidden reports whether err is a forbidden-type error.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden) ||
		errors.Is(err, ErrNotTeacher) ||
		errors.Is(err, ErrNotYourStudent) ||
		errors.Is(err, ErrAccessDenied)
}
