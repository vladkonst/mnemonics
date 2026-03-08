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
	UserID            int64
	ThemeID           int
	Status            Status
	Score             *int
	CurrentAttempt    int
	TestStartedAt     *time.Time
	StartedAt         time.Time
	CompletedAt       *time.Time
	TimeSpentSeconds  int
	LastViewedAt      *time.Time
	UpdatedAt         time.Time
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
