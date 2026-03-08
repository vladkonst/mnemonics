package user

import "fmt"

// SubscriptionStatus is a value object for the user's subscription state.
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusInactive SubscriptionStatus = "inactive"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
)

// NewSubscriptionStatus parses a subscription status string.
func NewSubscriptionStatus(s string) (SubscriptionStatus, error) {
	switch SubscriptionStatus(s) {
	case SubscriptionStatusActive, SubscriptionStatusInactive, SubscriptionStatusExpired:
		return SubscriptionStatus(s), nil
	default:
		return "", fmt.Errorf("invalid subscription status %q", s)
	}
}

func (s SubscriptionStatus) String() string { return string(s) }
