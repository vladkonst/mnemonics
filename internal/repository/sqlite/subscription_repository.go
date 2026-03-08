package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// SubscriptionRepo implements interfaces.SubscriptionRepository using SQLite.
type SubscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *subscription.Subscription) error {
	const q = `
		INSERT INTO subscriptions (
			payment_id, user_id, type, status, plan, expires_at,
			auto_renew, cancelled_at, cancellation_reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, q,
		s.PaymentID, s.UserID, string(s.Type), string(s.Status),
		s.Plan, s.ExpiresAt, boolToInt(s.AutoRenew),
		s.CancelledAt, s.CancellationReason,
	)
	return err
}

func (r *SubscriptionRepo) GetActiveByUserID(ctx context.Context, userID int64) (*subscription.Subscription, error) {
	const q = `
		SELECT payment_id, user_id, type, status, plan, expires_at,
		       auto_renew, cancelled_at, cancellation_reason, created_at
		FROM subscriptions
		WHERE user_id = ? AND status = 'active'
		ORDER BY created_at DESC LIMIT 1`

	row := r.db.QueryRowContext(ctx, q, userID)
	return scanSubscription(row)
}

func (r *SubscriptionRepo) GetByPaymentID(ctx context.Context, paymentID string) (*subscription.Subscription, error) {
	const q = `
		SELECT payment_id, user_id, type, status, plan, expires_at,
		       auto_renew, cancelled_at, cancellation_reason, created_at
		FROM subscriptions WHERE payment_id = ?`

	row := r.db.QueryRowContext(ctx, q, paymentID)
	return scanSubscription(row)
}

func scanSubscription(row *sql.Row) (*subscription.Subscription, error) {
	var s subscription.Subscription
	var typeStr, statusStr string
	var autoRenewInt int

	err := row.Scan(
		&s.PaymentID, &s.UserID, &typeStr, &statusStr,
		&s.Plan, &s.ExpiresAt, &autoRenewInt,
		&s.CancelledAt, &s.CancellationReason, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	s.Type = subscription.SubscriptionType(typeStr)
	s.Status = subscription.SubscriptionPlanStatus(statusStr)
	s.AutoRenew = autoRenewInt != 0
	return &s, nil
}
