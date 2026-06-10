package progress

import "time"

// ModuleTestQuestion is a question stored within a ModuleTestAttempt.
// It includes the correct answer so the attempt can be graded and reviewed later.
type ModuleTestQuestion struct {
	ID            int    `json:"id"`
	Text          string `json:"text"`
	CorrectAnswer string `json:"correct_answer"`
	OrderNum      int    `json:"order_num"`
}

// ModuleTestAttempt records a single dynamically-generated module test session.
// Questions are sampled on-the-fly from the theme tests belonging to the module.
type ModuleTestAttempt struct {
	ID              int                  `json:"id"`
	AttemptID       string               `json:"attempt_id"`
	UserID          int64                `json:"user_id"`
	ModuleID        int                  `json:"module_id"`
	Questions       []ModuleTestQuestion `json:"questions"`
	Answers         []AnswerItem         `json:"answers"`
	Score           int                  `json:"score"`
	Passed          bool                 `json:"passed"`
	StartedAt       time.Time            `json:"started_at"`
	SubmittedAt     *time.Time           `json:"submitted_at,omitempty"`
	DurationSeconds int                  `json:"duration_seconds"`
}

// IsSubmitted reports whether this attempt has already been graded.
func (a *ModuleTestAttempt) IsSubmitted() bool {
	return a.SubmittedAt != nil
}
