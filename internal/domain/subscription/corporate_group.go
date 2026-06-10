package subscription

import "time"

// CorporatePurchase records a manager's bulk group purchase.
type CorporatePurchase struct {
	PaymentID   string    `json:"payment_id"`
	ManagerID   int64     `json:"manager_id"`
	GroupsCount int       `json:"groups_count"`
	Semesters   int       `json:"semesters"`
	TotalAmount int       `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// CorporatePurchaseWithGroups contains a purchase and all its groups.
type CorporatePurchaseWithGroups struct {
	*CorporatePurchase
	Groups []*CorporateGroup `json:"groups"`
}

// CorporateGroup represents a single group within a corporate purchase.
type CorporateGroup struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	PurchaseID      string    `json:"purchase_id"`
	ManagerID       int64     `json:"manager_id"`
	TeacherID       *int64    `json:"teacher_id,omitempty"`
	TeacherJoinCode string    `json:"teacher_join_code"`
	StudentLinkID   string    `json:"student_link_id"`
	Semesters       int       `json:"semesters"`
	CreatedAt       time.Time `json:"created_at"`
}

// CorporateGroupWithStats includes student count statistics for a group.
type CorporateGroupWithStats struct {
	*CorporateGroup
	StudentCount int `json:"student_count"`
}
