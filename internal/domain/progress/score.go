package progress

import (
	"fmt"

	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// Score is a value object representing a test result percentage (0–100).
type Score struct {
	value int
}

// NewScore creates a Score, returning an error for out-of-range values.
func NewScore(v int) (Score, error) {
	if v < 0 || v > 100 {
		return Score{}, fmt.Errorf("%w: got %d", apperrors.ErrInvalidScore, v)
	}
	return Score{value: v}, nil
}

// Value returns the raw integer score.
func (s Score) Value() int { return s.value }

// Passed reports whether this score meets or exceeds the passing threshold.
func (s Score) Passed(passingScore int) bool { return s.value >= passingScore }

// Grade returns a Russian-style letter grade.
func (s Score) Grade() string {
	switch {
	case s.value >= 90:
		return "5"
	case s.value >= 75:
		return "4"
	case s.value >= 60:
		return "3"
	default:
		return "2"
	}
}
