package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// TestRepo implements interfaces.TestRepository using SQLite.
type TestRepo struct {
	db *sql.DB
}

func NewTestRepo(db *sql.DB) *TestRepo {
	return &TestRepo{db: db}
}

func (r *TestRepo) GetByThemeID(ctx context.Context, themeID int) (*content.Test, error) {
	const q = `
		SELECT id, theme_id, questions_json, difficulty, passing_score,
		       shuffle_questions, shuffle_answers, created_at
		FROM tests WHERE theme_id = ?`

	row := r.db.QueryRowContext(ctx, q, themeID)
	return scanTest(row)
}

func (r *TestRepo) GetByID(ctx context.Context, id int) (*content.Test, error) {
	const q = `
		SELECT id, theme_id, questions_json, difficulty, passing_score,
		       shuffle_questions, shuffle_answers, created_at
		FROM tests WHERE id = ?`

	row := r.db.QueryRowContext(ctx, q, id)
	return scanTest(row)
}

func (r *TestRepo) Create(ctx context.Context, t *content.Test) error {
	questionsJSON, err := json.Marshal(t.Questions)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO tests (theme_id, questions_json, difficulty, passing_score, shuffle_questions, shuffle_answers)
		VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, q,
		t.ThemeID, string(questionsJSON), t.Difficulty, t.PassingScore,
		boolToInt(t.ShuffleQuestions), boolToInt(t.ShuffleAnswers),
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = int(id)
	return nil
}

func scanTest(row *sql.Row) (*content.Test, error) {
	var t content.Test
	var questionsJSON string
	var shuffleQInt, shuffleAInt int

	err := row.Scan(
		&t.ID, &t.ThemeID, &questionsJSON, &t.Difficulty, &t.PassingScore,
		&shuffleQInt, &shuffleAInt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(questionsJSON), &t.Questions); err != nil {
		return nil, err
	}

	t.ShuffleQuestions = shuffleQInt != 0
	t.ShuffleAnswers = shuffleAInt != 0
	return &t, nil
}
