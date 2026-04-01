package subscription

import (
	"time"

	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// PromoCodeStatus is the lifecycle state of a PromoCode.
type PromoCodeStatus string

const (
	PromoCodeStatusPending     PromoCodeStatus = "pending"
	PromoCodeStatusActive      PromoCodeStatus = "active"
	PromoCodeStatusExpired     PromoCodeStatus = "expired"
	PromoCodeStatusDeactivated PromoCodeStatus = "deactivated"
)

// PromoCode is an aggregate root with its own identity and lifecycle.
// Lifecycle: pending → active → expired | deactivated
type PromoCode struct {
	Code             string          `json:"code"`
	UniversityName   string          `json:"university_name"`
	TeacherID        *int64          `json:"teacher_id,omitempty"`
	MaxActivations   int             `json:"max_activations"`
	Remaining        int             `json:"remaining"`
	Status           PromoCodeStatus `json:"status"`
	ExpiresAt        *time.Time      `json:"expires_at,omitempty"`
	CreatedByAdminID *int64          `json:"created_by_admin_id,omitempty"`
	ActivatedAt      *time.Time      `json:"activated_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

// Activate assigns the promo code to a teacher (pending → active).
func (p *PromoCode) Activate(teacherID int64) error {
	if p.TeacherID != nil {
		return apperrors.ErrAlreadyActivated
	}
	if p.Status != PromoCodeStatusPending {
		return apperrors.ErrInvalidStatus
	}
	now := time.Now().UTC()
	p.TeacherID = &teacherID
	p.Status = PromoCodeStatusActive
	p.ActivatedAt = &now
	return nil
}

// IsValidForStudent checks whether a student can use this promo code.
func (p *PromoCode) IsValidForStudent() error {
	if p.Status != PromoCodeStatusActive {
		return apperrors.ErrPromoCodeNotActive
	}
	if p.ExpiresAt != nil && time.Now().UTC().After(*p.ExpiresAt) {
		return apperrors.ErrPromoCodeExpired
	}
	if p.Remaining <= 0 {
		return apperrors.ErrPromoCodeExhausted
	}
	return nil
}

// Consume decrements remaining activations after a student joins.
func (p *PromoCode) Consume() error {
	if err := p.IsValidForStudent(); err != nil {
		return err
	}
	p.Remaining--
	return nil
}

// Deactivate marks the promo code as deactivated by an admin.
func (p *PromoCode) Deactivate() {
	p.Status = PromoCodeStatusDeactivated
}
