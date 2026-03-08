package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// UserRepo implements interfaces.UserRepository using SQLite.
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (
			telegram_id, role, subscription_status, university_code,
			pending_payment_id, first_name, last_name, username,
			language, timezone, notifications_enabled, last_activity_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, q,
		u.TelegramID,
		string(u.Role),
		string(u.SubscriptionStatus),
		u.UniversityCode,
		u.PendingPaymentID,
		u.FirstName,
		u.LastName,
		u.Username,
		u.Language,
		u.Timezone,
		boolToInt(u.NotificationsEnabled),
		u.LastActivityAt,
	)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, telegramID int64) (*user.User, error) {
	const q = `
		SELECT telegram_id, role, subscription_status, university_code,
		       pending_payment_id, first_name, last_name, username,
		       language, timezone, notifications_enabled, last_activity_at, created_at
		FROM users WHERE telegram_id = ?`

	row := r.db.QueryRowContext(ctx, q, telegramID)
	return scanUser(row)
}

func (r *UserRepo) Update(ctx context.Context, u *user.User) error {
	const q = `
		UPDATE users SET
			role = ?, subscription_status = ?, university_code = ?,
			pending_payment_id = ?, first_name = ?, last_name = ?, username = ?,
			language = ?, timezone = ?, notifications_enabled = ?, last_activity_at = ?
		WHERE telegram_id = ?`

	_, err := r.db.ExecContext(ctx, q,
		string(u.Role),
		string(u.SubscriptionStatus),
		u.UniversityCode,
		u.PendingPaymentID,
		u.FirstName,
		u.LastName,
		u.Username,
		u.Language,
		u.Timezone,
		boolToInt(u.NotificationsEnabled),
		u.LastActivityAt,
		u.TelegramID,
	)
	return err
}

func (r *UserRepo) Exists(ctx context.Context, telegramID int64) (bool, error) {
	const q = `SELECT COUNT(1) FROM users WHERE telegram_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, q, telegramID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// scanUser scans a single user row.
func scanUser(row *sql.Row) (*user.User, error) {
	var u user.User
	var roleStr, statusStr string
	var notifInt int

	err := row.Scan(
		&u.TelegramID,
		&roleStr,
		&statusStr,
		&u.UniversityCode,
		&u.PendingPaymentID,
		&u.FirstName,
		&u.LastName,
		&u.Username,
		&u.Language,
		&u.Timezone,
		&notifInt,
		&u.LastActivityAt,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	u.Role = user.Role(roleStr)
	u.SubscriptionStatus = user.SubscriptionStatus(statusStr)
	u.NotificationsEnabled = notifInt != 0
	return &u, nil
}

// boolToInt converts a bool to SQLite integer representation.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
