package content

import (
	"fmt"
	"strings"
	"time"
)

// ModuleTest is a test that covers an entire module, composed from questions
// across that module's themes. It mirrors the Test structure but is scoped
// to a module rather than a single theme.
type ModuleTest struct {
	ID               int        `json:"id"`
	ModuleID         int        `json:"module_id"`
	Name             string     `json:"name"`
	Questions        []Question `json:"questions"`
	Difficulty       int        `json:"difficulty"`
	PassingScore     int        `json:"passing_score"`
	ShuffleQuestions bool       `json:"shuffle_questions"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Validate checks structural integrity of the module test.
func (mt *ModuleTest) Validate() error {
	if mt.ModuleID <= 0 {
		return fmt.Errorf("module_id is required")
	}
	if mt.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(mt.Questions) == 0 {
		return fmt.Errorf("module test must have at least one question")
	}
	if mt.PassingScore < 0 || mt.PassingScore > 100 {
		return fmt.Errorf("passing_score must be 0–100, got %d", mt.PassingScore)
	}
	if mt.Difficulty < 1 || mt.Difficulty > 5 {
		return fmt.Errorf("difficulty must be 1–5, got %d", mt.Difficulty)
	}
	return nil
}

// Grade evaluates a set of answers and returns the score percentage and correct count.
// Comparison is case-insensitive.
func (mt *ModuleTest) Grade(answers map[int]string) (score int, correct int) {
	if len(mt.Questions) == 0 {
		return 0, 0
	}
	for _, q := range mt.Questions {
		if strings.EqualFold(strings.TrimSpace(answers[q.ID]), strings.TrimSpace(q.CorrectAnswer)) {
			correct++
		}
	}
	score = correct * 100 / len(mt.Questions)
	return score, correct
}

// Passed reports whether the given score meets the passing threshold.
func (mt *ModuleTest) Passed(score int) bool {
	return score >= mt.PassingScore
}
