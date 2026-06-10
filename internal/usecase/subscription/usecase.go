// Package subscription provides use cases for subscription management.
package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// UseCase orchestrates subscription operations.
type UseCase struct {
	subscriptions   interfaces.SubscriptionRepository
	users           interfaces.UserRepository
	teacherStudents interfaces.TeacherStudentRepository
	inviteLinks     interfaces.InviteLinkRepository
	notifications   interfaces.NotificationService
	corporateGroups interfaces.CorporateGroupRepository
}

// NewUseCase creates a new subscription UseCase.
func NewUseCase(
	subscriptions interfaces.SubscriptionRepository,
	users interfaces.UserRepository,
	teacherStudents interfaces.TeacherStudentRepository,
	inviteLinks interfaces.InviteLinkRepository,
	notifications interfaces.NotificationService,
	corporateGroups interfaces.CorporateGroupRepository,
) *UseCase {
	return &UseCase{
		subscriptions:   subscriptions,
		users:           users,
		teacherStudents: teacherStudents,
		inviteLinks:     inviteLinks,
		notifications:   notifications,
		corporateGroups: corporateGroups,
	}
}

// CreatePaymentSubscription activates a subscription after a successful payment.
func (uc *UseCase) CreatePaymentSubscription(ctx context.Context, userID int64, paymentID, plan string) (*subscription.Subscription, error) {
	existing, err := uc.subscriptions.GetByPaymentID(ctx, paymentID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	active, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if active != nil && active.IsActive() {
		return nil, apperrors.ErrActiveSubscriptionExists
	}

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

	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.ActivateSubscription(nil)
	u.ClearPendingPayment()
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	_ = uc.notifications.Send(ctx, userID, "Подписка успешно активирована! Приятного обучения.")
	return sub, nil
}

// GetTeacherInviteLinks returns all invite links for a teacher.
func (uc *UseCase) GetTeacherInviteLinks(ctx context.Context, teacherID int64) ([]*subscription.InviteLink, error) {
	return uc.inviteLinks.GetByTeacherID(ctx, teacherID)
}

// CreateInviteSubscription activates a subscription for a student via an invite link.
// Checks quota (max_activations), records activation, and links teacher↔student.
func (uc *UseCase) CreateInviteSubscription(ctx context.Context, userID int64, linkID string) (*subscription.Subscription, error) {
	link, err := uc.inviteLinks.GetByID(ctx, linkID)
	if err != nil {
		return nil, err
	}

	// Quota check.
	count, err := uc.inviteLinks.CountActivations(ctx, linkID)
	if err != nil {
		return nil, err
	}
	if count >= link.MaxActivations {
		return nil, apperrors.ErrInviteLinkExhausted
	}

	// No active subscription already.
	existing, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if existing != nil && existing.IsActive() {
		return nil, apperrors.ErrActiveSubscriptionExists
	}

	// Determine expiry from corporate group if applicable.
	var expiresAt *time.Time
	if uc.corporateGroups != nil {
		group, groupErr := uc.corporateGroups.GetByStudentLinkID(ctx, link.ID)
		if groupErr == nil && group != nil {
			t := time.Now().UTC().AddDate(0, group.Semesters*5, 0)
			expiresAt = &t
		}
	}

	// Activate user subscription.
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u.Role == user.RoleUnknown {
		u.SetRole(user.RoleStudent)
	}
	u.ActivateSubscription(nil)
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	// Create subscription record.
	now := time.Now().UTC()
	paymentID := fmt.Sprintf("invite-%s-%d-%d", linkID, userID, now.UnixNano())
	sub := &subscription.Subscription{
		PaymentID: paymentID,
		UserID:    userID,
		Type:      subscription.SubscriptionTypeUniversity,
		Status:    subscription.SubscriptionPlanStatusActive,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
	if err := uc.subscriptions.Create(ctx, sub); err != nil {
		return nil, err
	}

	// Link teacher↔student using invite link ID as join reference.
	if err := uc.teacherStudents.AddStudent(ctx, link.TeacherID, userID, link.ID); err != nil {
		return nil, err
	}

	// Record which link was used.
	act := &subscription.InviteLinkActivation{
		InviteLinkID: link.ID,
		UserID:       userID,
		ActivatedAt:  now,
	}
	_ = uc.inviteLinks.AddActivation(ctx, act)

	_ = uc.notifications.Send(ctx, userID, "Доступ по ссылке-приглашению активирован! Добро пожаловать.")
	return sub, nil
}
