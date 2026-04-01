package user

import (
	"testing"
	"time"
)

func TestNewRole(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
	}{
		{"student", false},
		{"teacher", false},
		{"admin", true},
		{"", true},
	}
	for _, c := range cases {
		r, err := NewRole(c.input)
		if c.wantErr {
			if err == nil {
				t.Errorf("NewRole(%q): expected error, got nil", c.input)
			}
		} else {
			if err != nil {
				t.Errorf("NewRole(%q): unexpected error: %v", c.input, err)
			}
			if r.String() != c.input {
				t.Errorf("NewRole(%q): got %q", c.input, r.String())
			}
		}
	}
}

func TestNewSubscriptionStatus(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
	}{
		{"active", false},
		{"inactive", false},
		{"expired", false},
		{"pending", true},
		{"", true},
	}
	for _, c := range cases {
		s, err := NewSubscriptionStatus(c.input)
		if c.wantErr {
			if err == nil {
				t.Errorf("NewSubscriptionStatus(%q): expected error", c.input)
			}
		} else {
			if err != nil {
				t.Errorf("NewSubscriptionStatus(%q): unexpected error: %v", c.input, err)
			}
			if s.String() != c.input {
				t.Errorf("NewSubscriptionStatus(%q): got %q", c.input, s.String())
			}
		}
	}
}

func TestUser_IsTeacher(t *testing.T) {
	student := &User{Role: RoleStudent}
	teacher := &User{Role: RoleTeacher}

	if student.IsTeacher() {
		t.Error("student.IsTeacher() should be false")
	}
	if !teacher.IsTeacher() {
		t.Error("teacher.IsTeacher() should be true")
	}
}

func TestUser_HasActiveSubscription(t *testing.T) {
	u := &User{SubscriptionStatus: SubscriptionStatusInactive}
	if u.HasActiveSubscription() {
		t.Error("inactive user should not have active subscription")
	}
	u.SubscriptionStatus = SubscriptionStatusActive
	if !u.HasActiveSubscription() {
		t.Error("active user should have active subscription")
	}
}

func TestUser_SetRole(t *testing.T) {
	u := &User{Role: RoleStudent}
	u.SetRole(RoleTeacher)
	if u.Role != RoleTeacher {
		t.Errorf("expected RoleTeacher, got %q", u.Role)
	}
}

func TestUser_ActivateSubscription(t *testing.T) {
	u := &User{SubscriptionStatus: SubscriptionStatusInactive}

	// without university code
	u.ActivateSubscription(nil)
	if u.SubscriptionStatus != SubscriptionStatusActive {
		t.Errorf("expected active, got %q", u.SubscriptionStatus)
	}
	if u.UniversityCode != nil {
		t.Error("university code should remain nil")
	}

	// with university code
	code := "MSU2024"
	u.ActivateSubscription(&code)
	if u.UniversityCode == nil || *u.UniversityCode != code {
		t.Errorf("expected university code %q", code)
	}
}

func TestUser_DeactivateSubscription(t *testing.T) {
	u := &User{SubscriptionStatus: SubscriptionStatusActive}
	u.DeactivateSubscription()
	if u.SubscriptionStatus != SubscriptionStatusExpired {
		t.Errorf("expected expired, got %q", u.SubscriptionStatus)
	}
}

func TestUser_SetAndClearPendingPayment(t *testing.T) {
	u := &User{}

	u.SetPendingPayment("pay_123")
	if u.PendingPaymentID == nil || *u.PendingPaymentID != "pay_123" {
		t.Error("expected pending payment ID to be set")
	}

	u.ClearPendingPayment()
	if u.PendingPaymentID != nil {
		t.Error("expected pending payment ID to be cleared")
	}
}

func TestUser_LastActivityAt(t *testing.T) {
	u := &User{}
	if u.LastActivityAt != nil {
		t.Error("LastActivityAt should be nil by default")
	}
	now := time.Now()
	u.LastActivityAt = &now
	if u.LastActivityAt == nil {
		t.Error("LastActivityAt should be set")
	}
}
