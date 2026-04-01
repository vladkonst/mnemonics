package content

import "time"

// Theme belongs to a Module and contains learning content.
// The first theme of each module (order_num=1) must be an introduction.
type Theme struct {
	ID                   int       `json:"id"`
	ModuleID             int       `json:"module_id"`
	Name                 string    `json:"name"`
	Description          *string   `json:"description,omitempty"`
	OrderNum             int       `json:"order_num"`
	IsIntroduction       bool      `json:"is_introduction"`
	IsLocked             bool      `json:"is_locked"`
	EstimatedTimeMinutes *int      `json:"estimated_time_minutes,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}
