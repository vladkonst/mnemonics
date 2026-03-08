package content

import "time"

// Theme belongs to a Module and contains learning content.
// The first theme of each module (order_num=1) must be an introduction.
type Theme struct {
	ID                   int
	ModuleID             int
	Name                 string
	Description          *string
	OrderNum             int
	IsIntroduction       bool
	IsLocked             bool
	EstimatedTimeMinutes *int
	CreatedAt            time.Time
}
