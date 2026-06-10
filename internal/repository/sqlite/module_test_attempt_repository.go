package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ModuleTestAttemptRepo implements interfaces.ModuleTestAttemptRepository using SQLite.
type ModuleTestAttemptRepo struct {
	db *sql.DB
}

func NewModuleTestAttemptRepo(db *sql.DB) *ModuleTestAttemptRepo {
	return &ModuleTestAttemptRepo{db: db}
}

func (r *ModuleTestAttemptRepo) Create(ctx context.Context, a *progress.ModuleTestAttempt) error {
	qJSON, err := json.Marshal(a.Questions)
	if err != nil {
		return err
	}
	aJSON, err := json.Marshal(a.Answers)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO module_test_attempts
			(attempt_id, user_id, module_id, questions_json, answers_json,
			 score, passed, started_at, submitted_at, duration_seconds)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q,
		a.AttemptID, a.UserID, a.ModuleID, string(qJSON), string(aJSON),
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

func (r *ModuleTestAttemptRepo) Update(ctx context.Context, a *progress.ModuleTestAttempt) error {
	aJSON, err := json.Marshal(a.Answers)
	if err != nil {
		return err
	}
	const q = `
		UPDATE module_test_attempts
		SET answers_json = ?, score = ?, passed = ?, submitted_at = ?, duration_seconds = ?
		WHERE attempt_id = ?`
	res, err := r.db.ExecContext(ctx, q,
		string(aJSON), a.Score, boolToInt(a.Passed), a.SubmittedAt, a.DurationSeconds, a.AttemptID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *ModuleTestAttemptRepo) GetByAttemptID(ctx context.Context, attemptID string) (*progress.ModuleTestAttempt, error) {
	const q = `
		SELECT id, attempt_id, user_id, module_id, questions_json, answers_json,
		       score, passed, started_at, submitted_at, duration_seconds
		FROM module_test_attempts WHERE attempt_id = ?`
	row := r.db.QueryRowContext(ctx, q, attemptID)
	var a progress.ModuleTestAttempt
	var qJSON, aJSON string
	var passedInt int
	err := row.Scan(&a.ID, &a.AttemptID, &a.UserID, &a.ModuleID,
		&qJSON, &aJSON, &a.Score, &passedInt, &a.StartedAt, &a.SubmittedAt, &a.DurationSeconds)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(qJSON), &a.Questions); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(aJSON), &a.Answers); err != nil {
		return nil, err
	}
	a.Passed = passedInt != 0
	return &a, nil
}

func (r *ModuleTestAttemptRepo) GetByUserAndModule(ctx context.Context, userID int64, moduleID int) ([]*progress.ModuleTestAttempt, error) {
	const q = `
		SELECT id, attempt_id, user_id, module_id, questions_json, answers_json,
		       score, passed, started_at, submitted_at, duration_seconds
		FROM module_test_attempts WHERE user_id = ? AND module_id = ? ORDER BY started_at DESC`
	rows, err := r.db.QueryContext(ctx, q, userID, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*progress.ModuleTestAttempt
	for rows.Next() {
		var a progress.ModuleTestAttempt
		var qJSON, aJSON string
		var passedInt int
		if err := rows.Scan(&a.ID, &a.AttemptID, &a.UserID, &a.ModuleID,
			&qJSON, &aJSON, &a.Score, &passedInt, &a.StartedAt, &a.SubmittedAt, &a.DurationSeconds); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(qJSON), &a.Questions); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(aJSON), &a.Answers); err != nil {
			return nil, err
		}
		a.Passed = passedInt != 0
		list = append(list, &a)
	}
	return list, rows.Err()
}
