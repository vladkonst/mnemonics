package content

import "time"

// Module is the top-level content organisation unit (aggregate root).
type Module struct {
	ID          int
	Name        string
	Description *string
	OrderNum    int
	IsLocked    bool
	IconEmoji   *string
	CreatedAt   time.Time
}
