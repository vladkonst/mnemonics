package content

import "time"

// Module is the top-level content organisation unit (aggregate root).
type Module struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OrderNum    int       `json:"order_num"`
	IsLocked    bool      `json:"is_locked"`
	IconEmoji   *string   `json:"icon_emoji,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
