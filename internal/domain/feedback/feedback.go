package feedback

import "time"

// Feedback is a message left by a user with suggestions or comments.
type Feedback struct {
	ID        int       `json:"id"`
	UserID    int64     `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
