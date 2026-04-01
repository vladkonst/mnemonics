package subscription

import (
	"testing"
	"time"
)

func TestSubscription_IsActive(t *testing.T) {
	// Active status, no expiry → active
	s := &Subscription{Status: SubscriptionPlanStatusActive}
	if !s.IsActive() {
		t.Error("active subscription with no expiry should be active")
	}

	// Active status, future expiry → active
	future := time.Now().UTC().Add(24 * time.Hour)
	s.ExpiresAt = &future
	if !s.IsActive() {
		t.Error("active subscription with future expiry should be active")
	}

	// Active status, past expiry → inactive
	past := time.Now().UTC().Add(-time.Hour)
	s.ExpiresAt = &past
	if s.IsActive() {
		t.Error("active subscription with past expiry should not be active")
	}

	// Expired status → inactive regardless of expiry
	s.ExpiresAt = nil
	s.Status = SubscriptionPlanStatusExpired
	if s.IsActive() {
		t.Error("expired status subscription should not be active")
	}

	// Cancelled status → inactive
	s.Status = SubscriptionPlanStatusCancelled
	if s.IsActive() {
		t.Error("cancelled subscription should not be active")
	}
}
