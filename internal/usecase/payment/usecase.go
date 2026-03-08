// Package payment provides use cases for payment invoice creation and webhook handling.
package payment

import (
	"context"
	"fmt"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// InvoiceResult is returned after creating a payment invoice.
type InvoiceResult struct {
	InvoiceID  string
	PaymentURL string
	Amount     int
	Plan       string
}

// WebhookEvent represents a parsed event from the payment gateway.
type WebhookEvent struct {
	PaymentID string
	UserID    int64
	Plan      string
	Status    string // "succeeded", "cancelled", etc.
}

// UseCase orchestrates payment operations.
type UseCase struct {
	users         interfaces.UserRepository
	subscriptions interfaces.SubscriptionRepository
	payment       interfaces.PaymentService
	notifications interfaces.NotificationService
}

// NewUseCase creates a new payment UseCase.
func NewUseCase(
	users interfaces.UserRepository,
	subscriptions interfaces.SubscriptionRepository,
	payment interfaces.PaymentService,
	notifications interfaces.NotificationService,
) *UseCase {
	return &UseCase{
		users:         users,
		subscriptions: subscriptions,
		payment:       payment,
		notifications: notifications,
	}
}

// CreateInvoice calls the payment gateway and records a pending payment on the user.
func (uc *UseCase) CreateInvoice(ctx context.Context, userID int64, plan string) (*InvoiceResult, error) {
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check for existing active subscription.
	active, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if active != nil && active.IsActive() {
		return nil, apperrors.ErrActiveSubscriptionExists
	}

	invoiceID, paymentURL, amount, err := uc.payment.CreateInvoice(ctx, userID, plan)
	if err != nil {
		return nil, fmt.Errorf("payment gateway: %w", err)
	}

	// Record the pending payment on the user.
	u.SetPendingPayment(invoiceID)
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	return &InvoiceResult{
		InvoiceID:  invoiceID,
		PaymentURL: paymentURL,
		Amount:     amount,
		Plan:       plan,
	}, nil
}

// HandleWebhook verifies the signature and idempotently processes a payment event.
func (uc *UseCase) HandleWebhook(ctx context.Context, payload []byte, signature string, event WebhookEvent) error {
	// Verify signature first.
	if err := uc.payment.VerifyWebhookSignature(payload, signature); err != nil {
		return fmt.Errorf("%w: %v", apperrors.ErrForbidden, err)
	}

	if event.Status != "succeeded" {
		// Nothing to do for non-success events.
		return nil
	}

	// Idempotency: check if subscription already created for this payment.
	existing, err := uc.subscriptions.GetByPaymentID(ctx, event.PaymentID)
	if err != nil && !apperrors.IsNotFound(err) {
		return err
	}
	if existing != nil {
		// Already processed.
		return nil
	}

	// Determine plan duration.
	plan := event.Plan
	if plan == "" {
		plan = "monthly"
	}

	// Activate subscription via CreatePaymentSubscription-like logic inline.
	// We do it here directly to avoid circular use case dependencies.
	u, err := uc.users.GetByID(ctx, event.UserID)
	if err != nil {
		return err
	}

	planCopy := plan
	sub := &subscription.Subscription{
		PaymentID: event.PaymentID,
		UserID:    event.UserID,
		Type:      subscription.SubscriptionTypePersonal,
		Status:    subscription.SubscriptionPlanStatusActive,
		Plan:      &planCopy,
		AutoRenew: false,
	}
	if err := uc.subscriptions.Create(ctx, sub); err != nil {
		return err
	}

	u.ActivateSubscription(nil)
	u.ClearPendingPayment()
	if err := uc.users.Update(ctx, u); err != nil {
		return err
	}

	// Notify user.
	_ = uc.notifications.Send(ctx, event.UserID, "Оплата прошла успешно! Подписка активирована.")

	return nil
}
