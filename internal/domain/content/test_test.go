package content_test

import (
	"testing"

	"github.com/vladkonst/mnemonics/internal/domain/content"
)

func makeTest(passingScore int, questions []content.Question) *content.Test {
	return &content.Test{
		ID:           1,
		ThemeID:      1,
		Difficulty:   1,
		PassingScore: passingScore,
		Questions:    questions,
	}
}

func TestTest_Grade(t *testing.T) {
	questions := []content.Question{
		{ID: 1, CorrectAnswer: "A"},
		{ID: 2, CorrectAnswer: "B"},
		{ID: 3, CorrectAnswer: "C"},
		{ID: 4, CorrectAnswer: "D"},
	}
	tst := makeTest(75, questions)

	t.Run("all correct", func(t *testing.T) {
		score, correct := tst.Grade(map[int]string{1: "A", 2: "B", 3: "C", 4: "D"})
		if score != 100 || correct != 4 {
			t.Errorf("got score=%d correct=%d, want 100/4", score, correct)
		}
	})

	t.Run("half correct", func(t *testing.T) {
		score, correct := tst.Grade(map[int]string{1: "A", 2: "X", 3: "C", 4: "X"})
		if score != 50 || correct != 2 {
			t.Errorf("got score=%d correct=%d, want 50/2", score, correct)
		}
	})

	t.Run("none correct", func(t *testing.T) {
		score, correct := tst.Grade(map[int]string{1: "X", 2: "X", 3: "X", 4: "X"})
		if score != 0 || correct != 0 {
			t.Errorf("got score=%d correct=%d, want 0/0", score, correct)
		}
	})
}

func TestTest_Passed(t *testing.T) {
	tst := makeTest(70, nil)
	if !tst.Passed(70) {
		t.Error("expected 70 to pass with passing_score=70")
	}
	if !tst.Passed(100) {
		t.Error("expected 100 to pass")
	}
	if tst.Passed(69) {
		t.Error("expected 69 to fail with passing_score=70")
	}
}
