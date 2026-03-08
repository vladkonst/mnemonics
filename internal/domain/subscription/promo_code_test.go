package subscription_test

import (
	"testing"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

func pendingPromo() *subscription.PromoCode {
	return &subscription.PromoCode{
		Code:           "TEST123",
		UniversityName: "МГУ",
		MaxActivations: 10,
		Remaining:      10,
		Status:         subscription.PromoCodeStatusPending,
	}
}

func TestPromoCode_Activate(t *testing.T) {
	p := pendingPromo()
	teacherID := int64(42)

	if err := p.Activate(teacherID); err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	if p.Status != subscription.PromoCodeStatusActive {
		t.Errorf("expected status=active, got %s", p.Status)
	}
	if p.TeacherID == nil || *p.TeacherID != teacherID {
		t.Error("TeacherID not set correctly")
	}
}

func TestPromoCode_Activate_AlreadyActivated(t *testing.T) {
	p := pendingPromo()
	_ = p.Activate(1)
	if err := p.Activate(2); !apperrors.IsConflict(err) {
		// ErrAlreadyActivated isn't a conflict sentinel — check directly
		if err == nil {
			t.Fatal("expected error on double activate, got nil")
		}
	}
}

func TestPromoCode_Consume(t *testing.T) {
	p := pendingPromo()
	_ = p.Activate(1)

	if err := p.Consume(); err != nil {
		t.Fatalf("Consume() unexpected error: %v", err)
	}
	if p.Remaining != 9 {
		t.Errorf("expected remaining=9, got %d", p.Remaining)
	}
}

func TestPromoCode_Consume_Exhausted(t *testing.T) {
	p := pendingPromo()
	p.Remaining = 0
	_ = p.Activate(1)
	// Override to exhausted
	p.Remaining = 0

	if err := p.Consume(); err == nil {
		t.Fatal("expected error when remaining=0")
	}
}

func TestPromoCode_Expired(t *testing.T) {
	p := pendingPromo()
	_ = p.Activate(1)
	past := time.Now().Add(-time.Hour)
	p.ExpiresAt = &past

	if err := p.IsValidForStudent(); err == nil {
		t.Fatal("expected expired error")
	}
}
