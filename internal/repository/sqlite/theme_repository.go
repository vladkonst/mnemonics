package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ThemeRepo implements interfaces.ThemeRepository using SQLite.
type ThemeRepo struct {
	db *sql.DB
}

func NewThemeRepo(db *sql.DB) *ThemeRepo {
	return &ThemeRepo{db: db}
}

func (r *ThemeRepo) GetByModuleID(ctx context.Context, moduleID int) ([]*content.Theme, error) {
	const q = `
		SELECT id, module_id, name, description, order_num, is_introduction, is_locked,
		       estimated_time_minutes, created_at
		FROM themes WHERE module_id = ? ORDER BY order_num`

	rows, err := r.db.QueryContext(ctx, q, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var themes []*content.Theme
	for rows.Next() {
		t, err := scanThemeRows(rows)
		if err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	return themes, rows.Err()
}

func (r *ThemeRepo) GetByID(ctx context.Context, id int) (*content.Theme, error) {
	const q = `
		SELECT id, module_id, name, description, order_num, is_introduction, is_locked,
		       estimated_time_minutes, created_at
		FROM themes WHERE id = ?`

	row := r.db.QueryRowContext(ctx, q, id)
	return scanThemeRow(row)
}

func (r *ThemeRepo) Create(ctx context.Context, t *content.Theme) error {
	const q = `
		INSERT INTO themes (module_id, name, description, order_num, is_introduction, is_locked, estimated_time_minutes)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, q,
		t.ModuleID, t.Name, t.Description, t.OrderNum,
		boolToInt(t.IsIntroduction), boolToInt(t.IsLocked), t.EstimatedTimeMinutes,
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

func (r *ThemeRepo) Update(ctx context.Context, t *content.Theme) (*content.Theme, error) {
	const q = `
		UPDATE themes SET name = ?, description = ?, order_num = ?, is_locked = ?, estimated_time_minutes = ?
		WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q,
		t.Name, t.Description, t.OrderNum, boolToInt(t.IsLocked), t.EstimatedTimeMinutes, t.ID,
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

func (r *ThemeRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM themes WHERE id = ?", id)
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

func (r *ThemeRepo) GetMaxOrderNum(ctx context.Context, moduleID int) (int, error) {
	var n int
	row := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(order_num), 0) FROM themes WHERE module_id = ?`, moduleID)
	err := row.Scan(&n)
	return n, err
}

// GetPreviousTheme returns the theme with order_num = current theme's order_num - 1 in the same module.
func (r *ThemeRepo) GetPreviousTheme(ctx context.Context, themeID int) (*content.Theme, error) {
	const q = `
		SELECT t2.id, t2.module_id, t2.name, t2.description, t2.order_num,
		       t2.is_introduction, t2.is_locked, t2.estimated_time_minutes, t2.created_at
		FROM themes t1
		JOIN themes t2 ON t1.module_id = t2.module_id AND t2.order_num = t1.order_num - 1
		WHERE t1.id = ?`

	row := r.db.QueryRowContext(ctx, q, themeID)
	return scanThemeRow(row)
}

func scanThemeRow(row *sql.Row) (*content.Theme, error) {
	var t content.Theme
	var isIntroInt, isLockedInt int
	err := row.Scan(
		&t.ID, &t.ModuleID, &t.Name, &t.Description, &t.OrderNum,
		&isIntroInt, &isLockedInt, &t.EstimatedTimeMinutes, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	t.IsIntroduction = isIntroInt != 0
	t.IsLocked = isLockedInt != 0
	return &t, nil
}

func scanThemeRows(rows *sql.Rows) (*content.Theme, error) {
	var t content.Theme
	var isIntroInt, isLockedInt int
	err := rows.Scan(
		&t.ID, &t.ModuleID, &t.Name, &t.Description, &t.OrderNum,
		&isIntroInt, &isLockedInt, &t.EstimatedTimeMinutes, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.IsIntroduction = isIntroInt != 0
	t.IsLocked = isLockedInt != 0
	return &t, nil
}
