package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ModuleTestRepo implements interfaces.ModuleTestRepository using SQLite.
type ModuleTestRepo struct {
	db *sql.DB
}

func NewModuleTestRepo(db *sql.DB) *ModuleTestRepo {
	return &ModuleTestRepo{db: db}
}

func (r *ModuleTestRepo) GetAll(ctx context.Context) ([]*content.ModuleTest, error) {
	const q = `
		SELECT id, module_id, name, questions_json, difficulty, passing_score,
		       shuffle_questions, created_at
		FROM module_tests ORDER BY id`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*content.ModuleTest
	for rows.Next() {
		t, err := scanModuleTest(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *ModuleTestRepo) GetByModuleID(ctx context.Context, moduleID int) ([]*content.ModuleTest, error) {
	const q = `
		SELECT id, module_id, name, questions_json, difficulty, passing_score,
		       shuffle_questions, created_at
		FROM module_tests WHERE module_id = ? ORDER BY id`

	rows, err := r.db.QueryContext(ctx, q, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*content.ModuleTest
	for rows.Next() {
		t, err := scanModuleTest(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *ModuleTestRepo) GetByID(ctx context.Context, id int) (*content.ModuleTest, error) {
	const q = `
		SELECT id, module_id, name, questions_json, difficulty, passing_score,
		       shuffle_questions, created_at
		FROM module_tests WHERE id = ?`

	row := r.db.QueryRowContext(ctx, q, id)
	var t content.ModuleTest
	var qJSON string
	var shuffleQInt int

	err := row.Scan(&t.ID, &t.ModuleID, &t.Name, &qJSON, &t.Difficulty, &t.PassingScore,
		&shuffleQInt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(qJSON), &t.Questions); err != nil {
		return nil, err
	}
	t.ShuffleQuestions = shuffleQInt != 0
	return &t, nil
}

func (r *ModuleTestRepo) Create(ctx context.Context, t *content.ModuleTest) error {
	questionsJSON, err := json.Marshal(t.Questions)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO module_tests (module_id, name, questions_json, difficulty, passing_score, shuffle_questions)
		VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, q,
		t.ModuleID, t.Name, string(questionsJSON), t.Difficulty, t.PassingScore,
		boolToInt(t.ShuffleQuestions),
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

func (r *ModuleTestRepo) Update(ctx context.Context, t *content.ModuleTest) (*content.ModuleTest, error) {
	questionsJSON, err := json.Marshal(t.Questions)
	if err != nil {
		return nil, err
	}
	const q = `
		UPDATE module_tests SET name = ?, questions_json = ?, difficulty = ?, passing_score = ?,
		                        shuffle_questions = ?
		WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q,
		t.Name, string(questionsJSON), t.Difficulty, t.PassingScore,
		boolToInt(t.ShuffleQuestions), t.ID,
	)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, apperrors.ErrNotFound
	}
	return t, nil
}

func (r *ModuleTestRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM module_tests WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func scanModuleTest(rows *sql.Rows) (*content.ModuleTest, error) {
	var t content.ModuleTest
	var qJSON string
	var shuffleQInt int

	if err := rows.Scan(&t.ID, &t.ModuleID, &t.Name, &qJSON, &t.Difficulty, &t.PassingScore,
		&shuffleQInt, &t.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(qJSON), &t.Questions); err != nil {
		return nil, err
	}
	t.ShuffleQuestions = shuffleQInt != 0
	return &t, nil
}
