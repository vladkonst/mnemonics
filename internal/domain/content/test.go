package content

import (
	"fmt"
	"strings"
	"time"
)

// Question is a value object within a Test.
// The only supported format is a free-text answer (case-insensitive match).
type Question struct {
	ID            int    `json:"id"`
	Text          string `json:"text"`
	CorrectAnswer string `json:"correct_answer"`
	OrderNum      int    `json:"order_num"`
}

// Test is an aggregate root that holds a set of Questions for a Theme.
type Test struct {
	ID               int        `json:"id"`
	ThemeID          int        `json:"theme_id"`
	Questions        []Question `json:"questions"`
	Difficulty       int        `json:"difficulty"`
	PassingScore     int        `json:"passing_score"`
	ShuffleQuestions bool       `json:"shuffle_questions"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Grade evaluates a set of answers and returns the score percentage.
// answers is a map of question_id → submitted answer string.
// Comparison is case-insensitive.
func (t *Test) Grade(answers map[int]string) (score int, correct int) {
	if len(t.Questions) == 0 {
		return 0, 0
	}
	for _, q := range t.Questions {
		if strings.EqualFold(strings.TrimSpace(answers[q.ID]), strings.TrimSpace(q.CorrectAnswer)) {
			correct++
		}
	}
	score = correct * 100 / len(t.Questions)
	return score, correct
}

// Passed reports whether the given score meets the passing threshold.
func (t *Test) Passed(score int) bool {
	return score >= t.PassingScore
}

// Validate checks basic structural integrity of the test.
func (t *Test) Validate() error {
	if len(t.Questions) == 0 {
		return fmt.Errorf("test must have at least one question")
	}
	if t.PassingScore < 0 || t.PassingScore > 100 {
		return fmt.Errorf("passing_score must be 0–100, got %d", t.PassingScore)
	}
	if t.Difficulty < 1 || t.Difficulty > 5 {
		return fmt.Errorf("difficulty must be 1–5, got %d", t.Difficulty)
	}
	return nil
}
