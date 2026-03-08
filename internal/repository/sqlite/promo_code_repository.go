package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// PromoCodeRepo implements interfaces.PromoCodeRepository using SQLite.
type PromoCodeRepo struct {
	db *sql.DB
}

func NewPromoCodeRepo(db *sql.DB) *PromoCodeRepo {
	return &PromoCodeRepo{db: db}
}

func (r *PromoCodeRepo) GetByCode(ctx context.Context, code string) (*subscription.PromoCode, error) {
	const q = `
		SELECT code, university_name, teacher_id, max_activations, remaining, status,
		       expires_at, created_by_admin_id, activated_at, created_at
		FROM promo_codes WHERE code = ?`

	row := r.db.QueryRowContext(ctx, q, code)
	p, err := scanPromoCode(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PromoCodeRepo) Update(ctx context.Context, p *subscription.PromoCode) error {
	const q = `
		UPDATE promo_codes SET
			university_name = ?, teacher_id = ?, max_activations = ?,
			remaining = ?, status = ?, expires_at = ?,
			created_by_admin_id = ?, activated_at = ?
		WHERE code = ?`

	_, err := r.db.ExecContext(ctx, q,
		p.UniversityName, p.TeacherID, p.MaxActivations,
		p.Remaining, string(p.Status), p.ExpiresAt,
		p.CreatedByAdminID, p.ActivatedAt,
		p.Code,
	)
	return err
}

func (r *PromoCodeRepo) Create(ctx context.Context, p *subscription.PromoCode) error {
	const q = `
		INSERT INTO promo_codes (
			code, university_name, teacher_id, max_activations, remaining, status,
			expires_at, created_by_admin_id, activated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, q,
		p.Code, p.UniversityName, p.TeacherID, p.MaxActivations,
		p.Remaining, string(p.Status), p.ExpiresAt,
		p.CreatedByAdminID, p.ActivatedAt,
	)
	return err
}

func (r *PromoCodeRepo) Deactivate(ctx context.Context, code string) error {
	const q = `UPDATE promo_codes SET status = 'deactivated' WHERE code = ?`
	_, err := r.db.ExecContext(ctx, q, code)
	return err
}

func (r *PromoCodeRepo) GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.PromoCode, error) {
	const q = `
		SELECT code, university_name, teacher_id, max_activations, remaining, status,
		       expires_at, created_by_admin_id, activated_at, created_at
		FROM promo_codes WHERE teacher_id = ?`

	rows, err := r.db.QueryContext(ctx, q, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []*subscription.PromoCode
	for rows.Next() {
		p, err := scanPromoCodeRows(rows)
		if err != nil {
			return nil, err
		}
		codes = append(codes, p)
	}
	return codes, rows.Err()
}

func scanPromoCode(row *sql.Row) (*subscription.PromoCode, error) {
	var p subscription.PromoCode
	var statusStr string

	err := row.Scan(
		&p.Code, &p.UniversityName, &p.TeacherID, &p.MaxActivations,
		&p.Remaining, &statusStr, &p.ExpiresAt,
		&p.CreatedByAdminID, &p.ActivatedAt, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Status = subscription.PromoCodeStatus(statusStr)
	return &p, nil
}

func scanPromoCodeRows(rows *sql.Rows) (*subscription.PromoCode, error) {
	var p subscription.PromoCode
	var statusStr string

	err := rows.Scan(
		&p.Code, &p.UniversityName, &p.TeacherID, &p.MaxActivations,
		&p.Remaining, &statusStr, &p.ExpiresAt,
		&p.CreatedByAdminID, &p.ActivatedAt, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Status = subscription.PromoCodeStatus(statusStr)
	return &p, nil
}
