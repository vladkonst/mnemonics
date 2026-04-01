package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ModuleRepo implements interfaces.ModuleRepository using SQLite.
type ModuleRepo struct {
	db *sql.DB
}

func NewModuleRepo(db *sql.DB) *ModuleRepo {
	return &ModuleRepo{db: db}
}

func (r *ModuleRepo) GetAll(ctx context.Context) ([]*content.Module, error) {
	const q = `
		SELECT id, name, description, order_num, is_locked, icon_emoji, created_at
		FROM modules ORDER BY order_num`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []*content.Module
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

func (r *ModuleRepo) GetByID(ctx context.Context, id int) (*content.Module, error) {
	const q = `
		SELECT id, name, description, order_num, is_locked, icon_emoji, created_at
		FROM modules WHERE id = ?`

	row := r.db.QueryRowContext(ctx, q, id)
	var m content.Module
	var isLockedInt int
	err := row.Scan(
		&m.ID, &m.Name, &m.Description, &m.OrderNum,
		&isLockedInt, &m.IconEmoji, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	m.IsLocked = isLockedInt != 0
	return &m, nil
}

func (r *ModuleRepo) Create(ctx context.Context, m *content.Module) error {
	const q = `
		INSERT INTO modules (name, description, order_num, is_locked, icon_emoji)
		VALUES (?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, q,
		m.Name, m.Description, m.OrderNum, boolToInt(m.IsLocked), m.IconEmoji,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	m.ID = int(id)
	return nil
}

func (r *ModuleRepo) Update(ctx context.Context, m *content.Module) error {
	const q = `
		UPDATE modules SET name = ?, description = ?, order_num = ?, is_locked = ?, icon_emoji = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, q,
		m.Name, m.Description, m.OrderNum, boolToInt(m.IsLocked), m.IconEmoji, m.ID,
	)
	return err
}

func (r *ModuleRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM modules WHERE id = ?", id)
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

func (r *ModuleRepo) GetMaxOrderNum(ctx context.Context) (int, error) {
	var n int
	row := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(order_num), 0) FROM modules`)
	err := row.Scan(&n)
	return n, err
}

// scanModule scans a module from sql.Rows.
func scanModule(rows *sql.Rows) (*content.Module, error) {
	var m content.Module
	var isLockedInt int
	err := rows.Scan(
		&m.ID, &m.Name, &m.Description, &m.OrderNum,
		&isLockedInt, &m.IconEmoji, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.IsLocked = isLockedInt != 0
	return &m, nil
}
