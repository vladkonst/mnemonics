package progress

import "time"

// AnswerItem holds a single submitted answer.
type AnswerItem struct {
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
}

// TestAttempt records the history of a single test submission.
type TestAttempt struct {
	ID              int          `json:"id"`
	UserID          int64        `json:"user_id"`
	ThemeID         int          `json:"theme_id"`
	TestID          int          `json:"test_id"`
	AttemptID       string       `json:"attempt_id"`
	Answers         []AnswerItem `json:"answers"`
	Score           int          `json:"score"`
	Passed          bool         `json:"passed"`
	StartedAt       time.Time    `json:"started_at"`
	SubmittedAt     *time.Time   `json:"submitted_at,omitempty"`
	DurationSeconds int          `json:"duration_seconds"`
}

// IsSubmitted reports whether this attempt has already been scored.
func (a *TestAttempt) IsSubmitted() bool {
	return a.SubmittedAt != nil
}
