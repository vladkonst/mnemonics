package sqlite

import (
	"context"
	"database/sql"

	"github.com/vladkonst/mnemonics/internal/domain/feedback"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

type FeedbackRepo struct {
	db *sql.DB
}

func NewFeedbackRepo(db *sql.DB) *FeedbackRepo {
	return &FeedbackRepo{db: db}
}

func (r *FeedbackRepo) Create(ctx context.Context, f *feedback.Feedback) error {
	const q = `INSERT INTO feedback (user_id, text, created_at) VALUES (?, ?, ?) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, q, f.UserID, f.Text, f.CreatedAt).Scan(&f.ID, &f.CreatedAt)
}

func (r *FeedbackRepo) GetAll(ctx context.Context, limit, offset int) ([]*feedback.Feedback, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM feedback`).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `SELECT id, user_id, text, created_at FROM feedback ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*feedback.Feedback
	for rows.Next() {
		f := &feedback.Feedback{}
		if err := rows.Scan(&f.ID, &f.UserID, &f.Text, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, f)
	}
	return items, total, rows.Err()
}

func (r *FeedbackRepo) GetByID(ctx context.Context, id int) (*feedback.Feedback, error) {
	const q = `SELECT id, user_id, text, created_at FROM feedback WHERE id = ?`
	f := &feedback.Feedback{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(&f.ID, &f.UserID, &f.Text, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	return f, err
}
