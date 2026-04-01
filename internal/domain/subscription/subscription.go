package subscription

import "time"

// SubscriptionType indicates whether the subscription is personal or university.
type SubscriptionType string

const (
	SubscriptionTypePersonal   SubscriptionType = "personal"
	SubscriptionTypeUniversity SubscriptionType = "university"
)

// SubscriptionPlanStatus tracks payment subscription state.
type SubscriptionPlanStatus string

const (
	SubscriptionPlanStatusActive    SubscriptionPlanStatus = "active"
	SubscriptionPlanStatusExpired   SubscriptionPlanStatus = "expired"
	SubscriptionPlanStatusCancelled SubscriptionPlanStatus = "cancelled"
)

// Subscription records a paid or promo-granted access period.
type Subscription struct {
	PaymentID          string                 `json:"payment_id"`
	UserID             int64                  `json:"user_id"`
	Type               SubscriptionType       `json:"type"`
	Status             SubscriptionPlanStatus `json:"status"`
	Plan               *string                `json:"plan,omitempty"`
	ExpiresAt          *time.Time             `json:"expires_at,omitempty"`
	AutoRenew          bool                   `json:"auto_renew"`
	CancelledAt        *time.Time             `json:"cancelled_at,omitempty"`
	CancellationReason *string                `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

// IsActive reports whether the subscription grants current access.
func (s *Subscription) IsActive() bool {
	if s.Status != SubscriptionPlanStatusActive {
		return false
	}
	if s.ExpiresAt != nil && time.Now().UTC().After(*s.ExpiresAt) {
		return false
	}
	return true
}
