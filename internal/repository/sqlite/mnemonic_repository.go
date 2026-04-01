package sqlite

import (
	"context"
	"database/sql"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// MnemonicRepo implements interfaces.MnemonicRepository using SQLite.
type MnemonicRepo struct {
	db *sql.DB
}

func NewMnemonicRepo(db *sql.DB) *MnemonicRepo {
	return &MnemonicRepo{db: db}
}

func (r *MnemonicRepo) GetByThemeID(ctx context.Context, themeID int) ([]*content.Mnemonic, error) {
	const q = `
		SELECT id, theme_id, type, content_text, s3_image_key, order_num, created_at
		FROM mnemonics WHERE theme_id = ? ORDER BY order_num`

	rows, err := r.db.QueryContext(ctx, q, themeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mnemonics []*content.Mnemonic
	for rows.Next() {
		var m content.Mnemonic
		var typeStr string
		err := rows.Scan(
			&m.ID, &m.ThemeID, &typeStr,
			&m.ContentText, &m.S3ImageKey, &m.OrderNum, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		m.Type = content.MnemonicType(typeStr)
		mnemonics = append(mnemonics, &m)
	}
	return mnemonics, rows.Err()
}

func (r *MnemonicRepo) Create(ctx context.Context, m *content.Mnemonic) error {
	const q = `
		INSERT INTO mnemonics (theme_id, type, content_text, s3_image_key, order_num)
		VALUES (?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, q,
		m.ThemeID, string(m.Type), m.ContentText, m.S3ImageKey, m.OrderNum,
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

func (r *MnemonicRepo) GetMaxOrderNum(ctx context.Context, themeID int) (int, error) {
	var n int
	row := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(order_num), 0) FROM mnemonics WHERE theme_id = ?`, themeID)
	err := row.Scan(&n)
	return n, err
}

func (r *MnemonicRepo) Update(ctx context.Context, m *content.Mnemonic) (*content.Mnemonic, error) {
	const q = `UPDATE mnemonics SET content_text = ?, s3_image_key = ?, order_num = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, m.ContentText, m.S3ImageKey, m.OrderNum, m.ID)
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
	return m, nil
}

func (r *MnemonicRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM mnemonics WHERE id = ?", id)
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
