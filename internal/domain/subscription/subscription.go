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
	PaymentID          string
	UserID             int64
	Type               SubscriptionType
	Status             SubscriptionPlanStatus
	Plan               *string
	ExpiresAt          *time.Time
	AutoRenew          bool
	CancelledAt        *time.Time
	CancellationReason *string
	CreatedAt          time.Time
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
