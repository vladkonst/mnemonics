package progress

import "time"

// AnswerItem holds a single submitted answer.
type AnswerItem struct {
	QuestionID int
	Answer     string
}

// TestAttempt records the history of a single test submission.
type TestAttempt struct {
	ID              int
	UserID          int64
	ThemeID         int
	TestID          int
	AttemptID       string // UUID, used as idempotency key
	Answers         []AnswerItem
	Score           int
	Passed          bool
	StartedAt       time.Time
	SubmittedAt     *time.Time
	DurationSeconds int
}

// IsSubmitted reports whether this attempt has already been scored.
func (a *TestAttempt) IsSubmitted() bool {
	return a.SubmittedAt != nil
}
