package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ProgressRepo implements interfaces.ProgressRepository using SQLite.
type ProgressRepo struct {
	db *sql.DB
}

func NewProgressRepo(db *sql.DB) *ProgressRepo {
	return &ProgressRepo{db: db}
}

func (r *ProgressRepo) Upsert(ctx context.Context, p *progress.UserProgress) error {
	const q = `
		INSERT INTO user_progress (
			user_id, theme_id, status, score, current_attempt,
			test_started_at, started_at, completed_at, time_spent_seconds,
			last_viewed_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, theme_id) DO UPDATE SET
			status            = excluded.status,
			score             = excluded.score,
			current_attempt   = excluded.current_attempt,
			test_started_at   = excluded.test_started_at,
			completed_at      = excluded.completed_at,
			time_spent_seconds = excluded.time_spent_seconds,
			last_viewed_at    = excluded.last_viewed_at,
			updated_at        = CURRENT_TIMESTAMP`

	_, err := r.db.ExecContext(ctx, q,
		p.UserID, p.ThemeID, string(p.Status), p.Score, p.CurrentAttempt,
		p.TestStartedAt, p.StartedAt, p.CompletedAt, p.TimeSpentSeconds,
		p.LastViewedAt,
	)
	return err
}

func (r *ProgressRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) (*progress.UserProgress, error) {
	const q = `
		SELECT user_id, theme_id, status, score, current_attempt,
		       test_started_at, started_at, completed_at, time_spent_seconds,
		       last_viewed_at, updated_at
		FROM user_progress WHERE user_id = ? AND theme_id = ?`

	row := r.db.QueryRowContext(ctx, q, userID, themeID)
	return scanProgress(row)
}

func (r *ProgressRepo) GetByUser(ctx context.Context, userID int64) ([]*progress.UserProgress, error) {
	const q = `
		SELECT user_id, theme_id, status, score, current_attempt,
		       test_started_at, started_at, completed_at, time_spent_seconds,
		       last_viewed_at, updated_at
		FROM user_progress WHERE user_id = ? ORDER BY started_at`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProgressRows(rows)
}

func (r *ProgressRepo) GetByUserAndModule(ctx context.Context, userID int64, moduleID int) ([]*progress.UserProgress, error) {
	const q = `
		SELECT up.user_id, up.theme_id, up.status, up.score, up.current_attempt,
		       up.test_started_at, up.started_at, up.completed_at, up.time_spent_seconds,
		       up.last_viewed_at, up.updated_at
		FROM user_progress up
		JOIN themes t ON t.id = up.theme_id
		WHERE up.user_id = ? AND t.module_id = ?
		ORDER BY t.order_num`

	rows, err := r.db.QueryContext(ctx, q, userID, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProgressRows(rows)
}

func (r *ProgressRepo) CountCompletedByUser(ctx context.Context, userID int64) (int, error) {
	const q = `SELECT COUNT(*) FROM user_progress WHERE user_id = ? AND status = 'completed'`
	var count int
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&count)
	return count, err
}

func scanProgress(row *sql.Row) (*progress.UserProgress, error) {
	var p progress.UserProgress
	var statusStr string

	err := row.Scan(
		&p.UserID, &p.ThemeID, &statusStr, &p.Score, &p.CurrentAttempt,
		&p.TestStartedAt, &p.StartedAt, &p.CompletedAt, &p.TimeSpentSeconds,
		&p.LastViewedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	p.Status = progress.Status(statusStr)
	return &p, nil
}

func scanProgressRows(rows *sql.Rows) ([]*progress.UserProgress, error) {
	var result []*progress.UserProgress
	for rows.Next() {
		var p progress.UserProgress
		var statusStr string

		err := rows.Scan(
			&p.UserID, &p.ThemeID, &statusStr, &p.Score, &p.CurrentAttempt,
			&p.TestStartedAt, &p.StartedAt, &p.CompletedAt, &p.TimeSpentSeconds,
			&p.LastViewedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		p.Status = progress.Status(statusStr)
		result = append(result, &p)
	}
	return result, rows.Err()
}
