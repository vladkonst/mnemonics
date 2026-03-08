package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// TestAttemptRepo implements interfaces.TestAttemptRepository using SQLite.
type TestAttemptRepo struct {
	db *sql.DB
}

func NewTestAttemptRepo(db *sql.DB) *TestAttemptRepo {
	return &TestAttemptRepo{db: db}
}

func (r *TestAttemptRepo) Create(ctx context.Context, a *progress.TestAttempt) error {
	answersJSON, err := json.Marshal(a.Answers)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO test_attempts (
			user_id, theme_id, test_id, attempt_id, answers_json,
			score, passed, started_at, submitted_at, duration_seconds
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, q,
		a.UserID, a.ThemeID, a.TestID, a.AttemptID, string(answersJSON),
		a.Score, boolToInt(a.Passed), a.StartedAt, a.SubmittedAt, a.DurationSeconds,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = int(id)
	return nil
}

func (r *TestAttemptRepo) GetByAttemptID(ctx context.Context, attemptID string) (*progress.TestAttempt, error) {
	const q = `
		SELECT id, user_id, theme_id, test_id, attempt_id, answers_json,
		       score, passed, started_at, submitted_at, duration_seconds
		FROM test_attempts WHERE attempt_id = ?`

	row := r.db.QueryRowContext(ctx, q, attemptID)
	a, err := scanAttempt(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *TestAttemptRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) ([]*progress.TestAttempt, error) {
	const q = `
		SELECT id, user_id, theme_id, test_id, attempt_id, answers_json,
		       score, passed, started_at, submitted_at, duration_seconds
		FROM test_attempts
		WHERE user_id = ? AND theme_id = ?
		ORDER BY submitted_at DESC`

	rows, err := r.db.QueryContext(ctx, q, userID, themeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []*progress.TestAttempt
	for rows.Next() {
		a, err := scanAttemptRows(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}

func scanAttempt(row *sql.Row) (*progress.TestAttempt, error) {
	var a progress.TestAttempt
	var answersJSON string
	var passedInt int

	err := row.Scan(
		&a.ID, &a.UserID, &a.ThemeID, &a.TestID, &a.AttemptID,
		&answersJSON, &a.Score, &passedInt,
		&a.StartedAt, &a.SubmittedAt, &a.DurationSeconds,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(answersJSON), &a.Answers); err != nil {
		return nil, err
	}
	a.Passed = passedInt != 0
	return &a, nil
}

func scanAttemptRows(rows *sql.Rows) (*progress.TestAttempt, error) {
	var a progress.TestAttempt
	var answersJSON string
	var passedInt int

	err := rows.Scan(
		&a.ID, &a.UserID, &a.ThemeID, &a.TestID, &a.AttemptID,
		&answersJSON, &a.Score, &passedInt,
		&a.StartedAt, &a.SubmittedAt, &a.DurationSeconds,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(answersJSON), &a.Answers); err != nil {
		return nil, err
	}
	a.Passed = passedInt != 0
	return &a, nil
}
