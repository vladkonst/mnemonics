package progress_test

import (
	"testing"

	"github.com/vladkonst/mnemonics/internal/domain/progress"
)

func TestNewScore_Valid(t *testing.T) {
	for _, v := range []int{0, 50, 100} {
		s, err := progress.NewScore(v)
		if err != nil {
			t.Errorf("NewScore(%d) unexpected error: %v", v, err)
		}
		if s.Value() != v {
			t.Errorf("Value() = %d, want %d", s.Value(), v)
		}
	}
}

func TestNewScore_Invalid(t *testing.T) {
	for _, v := range []int{-1, 101} {
		if _, err := progress.NewScore(v); err == nil {
			t.Errorf("NewScore(%d) expected error, got nil", v)
		}
	}
}

func TestScore_Grade(t *testing.T) {
	cases := []struct {
		score    int
		expected string
	}{
		{90, "5"}, {95, "5"}, {100, "5"},
		{75, "4"}, {89, "4"},
		{60, "3"}, {74, "3"},
		{0, "2"}, {59, "2"},
	}
	for _, c := range cases {
		s, _ := progress.NewScore(c.score)
		if got := s.Grade(); got != c.expected {
			t.Errorf("Grade(%d) = %q, want %q", c.score, got, c.expected)
		}
	}
}
