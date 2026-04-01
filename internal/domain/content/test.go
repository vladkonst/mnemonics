package content

import (
	"fmt"
	"time"
)

// QuestionType defines supported question formats.
type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeTrueFalse      QuestionType = "true_false"
)

// Question is a value object within a Test.
type Question struct {
	ID            int          `json:"id"`
	Text          string       `json:"text"`
	Type          QuestionType `json:"type"`
	Options       []string     `json:"options"`
	CorrectAnswer string       `json:"correct_answer"`
	OrderNum      int          `json:"order_num"`
}

// Test is an aggregate root that holds a set of Questions for a Theme.
type Test struct {
	ID               int        `json:"id"`
	ThemeID          int        `json:"theme_id"`
	Questions        []Question `json:"questions"`
	Difficulty       int        `json:"difficulty"`
	PassingScore     int        `json:"passing_score"`
	ShuffleQuestions bool       `json:"shuffle_questions"`
	ShuffleAnswers   bool       `json:"shuffle_answers"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Grade evaluates a set of answers and returns the score percentage.
// answers is a map of question_id → submitted answer string.
func (t *Test) Grade(answers map[int]string) (score int, correct int) {
	if len(t.Questions) == 0 {
		return 0, 0
	}
	for _, q := range t.Questions {
		if answers[q.ID] == q.CorrectAnswer {
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
