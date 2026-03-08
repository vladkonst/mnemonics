// Package subscription provides use cases for promo code and subscription management.
package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// UseCase orchestrates subscription and promo code operations.
type UseCase struct {
	promoCodes      interfaces.PromoCodeRepository
	subscriptions   interfaces.SubscriptionRepository
	users           interfaces.UserRepository
	teacherStudents interfaces.TeacherStudentRepository
	notifications   interfaces.NotificationService
}

// NewUseCase creates a new subscription UseCase.
func NewUseCase(
	promoCodes interfaces.PromoCodeRepository,
	subscriptions interfaces.SubscriptionRepository,
	users interfaces.UserRepository,
	teacherStudents interfaces.TeacherStudentRepository,
	notifications interfaces.NotificationService,
) *UseCase {
	return &UseCase{
		promoCodes:      promoCodes,
		subscriptions:   subscriptions,
		users:           users,
		teacherStudents: teacherStudents,
		notifications:   notifications,
	}
}

// ActivatePromoCode assigns a pending promo code to a teacher (teacher claims the code).
func (uc *UseCase) ActivatePromoCode(ctx context.Context, teacherID int64, code string) (*subscription.PromoCode, error) {
	// Verify the teacher exists and is a teacher.
	u, err := uc.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	if !u.IsTeacher() {
		return nil, apperrors.ErrNotTeacher
	}

	promo, err := uc.promoCodes.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if err := promo.Activate(teacherID); err != nil {
		return nil, err
	}

	if err := uc.promoCodes.Update(ctx, promo); err != nil {
		return nil, err
	}
	return promo, nil
}

// CreatePromoSubscription allows a student to join via a promo code.
func (uc *UseCase) CreatePromoSubscription(ctx context.Context, userID int64, code string) (*subscription.Subscription, error) {
	// Check for existing active subscription.
	existing, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if existing != nil && existing.IsActive() {
		return nil, apperrors.ErrActiveSubscriptionExists
	}

	promo, err := uc.promoCodes.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if err := promo.Consume(); err != nil {
		return nil, err
	}

	if err := uc.promoCodes.Update(ctx, promo); err != nil {
		return nil, err
	}

	// Record teacher–student relationship.
	if promo.TeacherID != nil {
		if err := uc.teacherStudents.AddStudent(ctx, *promo.TeacherID, userID, code); err != nil {
			return nil, err
		}
	}

	// Activate the user's subscription.
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	uniCode := promo.Code
	u.ActivateSubscription(&uniCode)
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	// Create subscription record (no expiry for promo by default).
	now := time.Now().UTC()
	paymentID := fmt.Sprintf("promo-%s-%d-%d", code, userID, now.UnixNano())
	sub := &subscription.Subscription{
		PaymentID: paymentID,
		UserID:    userID,
		Type:      subscription.SubscriptionTypeUniversity,
		Status:    subscription.SubscriptionPlanStatusActive,
		CreatedAt: now,
	}
	if err := uc.subscriptions.Create(ctx, sub); err != nil {
		return nil, err
	}

	// Notify user.
	_ = uc.notifications.Send(ctx, userID, "Доступ по промокоду активирован! Добро пожаловать.")

	return sub, nil
}

// CreatePaymentSubscription activates a subscription after a successful payment.
func (uc *UseCase) CreatePaymentSubscription(ctx context.Context, userID int64, paymentID, plan string) (*subscription.Subscription, error) {
	// Idempotency: check if subscription for this payment already exists.
	existing, err := uc.subscriptions.GetByPaymentID(ctx, paymentID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Check no active subscription already.
	active, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if active != nil && active.IsActive() {
		return nil, apperrors.ErrActiveSubscriptionExists
	}

	// Determine expiry based on plan.
	now := time.Now().UTC()
	var expiresAt *time.Time
	switch plan {
	case "monthly":
		t := now.AddDate(0, 1, 0)
		expiresAt = &t
	case "yearly":
		t := now.AddDate(1, 0, 0)
		expiresAt = &t
	default:
		t := now.AddDate(0, 1, 0)
		expiresAt = &t
	}
	planCopy := plan

	sub := &subscription.Subscription{
		PaymentID: paymentID,
		UserID:    userID,
		Type:      subscription.SubscriptionTypePersonal,
		Status:    subscription.SubscriptionPlanStatusActive,
		Plan:      &planCopy,
		ExpiresAt: expiresAt,
		AutoRenew: false,
		CreatedAt: now,
	}
	if err := uc.subscriptions.Create(ctx, sub); err != nil {
		return nil, err
	}

	// Activate user subscription status.
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.ActivateSubscription(nil)
	u.ClearPendingPayment()
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	// Notify user.
	_ = uc.notifications.Send(ctx, userID, "Подписка успешно активирована! Приятного обучения.")

	return sub, nil
}

// GetTeacherPromoCodes returns all promo codes created by or assigned to a teacher.
func (uc *UseCase) GetTeacherPromoCodes(ctx context.Context, teacherID int64) ([]*subscription.PromoCode, error) {
	u, err := uc.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	if !u.IsTeacher() {
		return nil, apperrors.ErrNotTeacher
	}

	return uc.promoCodes.GetByTeacherID(ctx, teacherID)
}
