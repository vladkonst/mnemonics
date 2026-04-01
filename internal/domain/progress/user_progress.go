package progress

import "time"

// Status represents the state of a user's progress on a theme.
type Status string

const (
	StatusStarted   Status = "started"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// UserProgress is the aggregate root tracking a user's state for one theme.
// Primary key: (UserID, ThemeID).
type UserProgress struct {
	UserID           int64      `json:"user_id"`
	ThemeID          int        `json:"theme_id"`
	Status           Status     `json:"status"`
	Score            *int       `json:"score,omitempty"`
	CurrentAttempt   int        `json:"current_attempt"`
	TestStartedAt    *time.Time `json:"test_started_at,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	TimeSpentSeconds int        `json:"time_spent_seconds"`
	LastViewedAt     *time.Time `json:"last_viewed_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// MarkStarted records when a user begins studying a theme.
func (p *UserProgress) MarkStarted() {
	now := time.Now().UTC()
	p.Status = StatusStarted
	p.LastViewedAt = &now
	p.UpdatedAt = now
}

// StartTest records that a test attempt has begun.
func (p *UserProgress) StartTest() {
	now := time.Now().UTC()
	p.TestStartedAt = &now
	p.CurrentAttempt++
	p.UpdatedAt = now
}

// Complete marks the theme as successfully completed with the given score.
func (p *UserProgress) Complete(score int) {
	now := time.Now().UTC()
	p.Status = StatusCompleted
	p.Score = &score
	p.CompletedAt = &now
	p.UpdatedAt = now
}

// Fail marks the theme attempt as failed with the given score.
func (p *UserProgress) Fail(score int) {
	now := time.Now().UTC()
	p.Status = StatusFailed
	p.Score = &score
	p.UpdatedAt = now
}

// IsCompleted reports whether the theme has been successfully completed.
func (p *UserProgress) IsCompleted() bool {
	return p.Status == StatusCompleted
}
