package sqlite

import (
	"context"
	"database/sql"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// CorporateGroupRepo implements interfaces.CorporateGroupRepository using SQLite.
type CorporateGroupRepo struct {
	db *sql.DB
}

// NewCorporateGroupRepo creates a new CorporateGroupRepo.
func NewCorporateGroupRepo(db *sql.DB) *CorporateGroupRepo {
	return &CorporateGroupRepo{db: db}
}

// CreatePurchase inserts a new corporate purchase record.
func (r *CorporateGroupRepo) CreatePurchase(ctx context.Context, p *subscription.CorporatePurchase) error {
	const q = `INSERT INTO corporate_purchases (payment_id, manager_id, groups_count, semesters, total_amount, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, p.PaymentID, p.ManagerID, p.GroupsCount, p.Semesters, p.TotalAmount, p.CreatedAt)
	return err
}

// GetPurchaseByPaymentID fetches a single corporate purchase by its payment ID.
func (r *CorporateGroupRepo) GetPurchaseByPaymentID(ctx context.Context, paymentID string) (*subscription.CorporatePurchase, error) {
	const q = `SELECT payment_id, manager_id, groups_count, semesters, total_amount, created_at
		FROM corporate_purchases WHERE payment_id = ?`
	p := &subscription.CorporatePurchase{}
	err := r.db.QueryRowContext(ctx, q, paymentID).Scan(
		&p.PaymentID, &p.ManagerID, &p.GroupsCount, &p.Semesters, &p.TotalAmount, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	return p, err
}

// GetPurchasesByManagerID returns all corporate purchases for a manager.
func (r *CorporateGroupRepo) GetPurchasesByManagerID(ctx context.Context, managerID int64) ([]*subscription.CorporatePurchase, error) {
	const q = `SELECT payment_id, manager_id, groups_count, semesters, total_amount, created_at
		FROM corporate_purchases WHERE manager_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, managerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []*subscription.CorporatePurchase
	for rows.Next() {
		p := &subscription.CorporatePurchase{}
		if err := rows.Scan(&p.PaymentID, &p.ManagerID, &p.GroupsCount, &p.Semesters, &p.TotalAmount, &p.CreatedAt); err != nil {
			return nil, err
		}
		purchases = append(purchases, p)
	}
	return purchases, rows.Err()
}

// GetGroupsByPurchaseID returns all corporate groups for a given purchase.
func (r *CorporateGroupRepo) GetGroupsByPurchaseID(ctx context.Context, purchaseID string) ([]*subscription.CorporateGroup, error) {
	const q = `SELECT id, name, purchase_id, manager_id, teacher_id, teacher_join_code, student_link_id, semesters, created_at
		FROM corporate_groups WHERE purchase_id = ? ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, q, purchaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*subscription.CorporateGroup
	for rows.Next() {
		g, err := scanCorporateGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// Create inserts a new corporate group record.
func (r *CorporateGroupRepo) Create(ctx context.Context, g *subscription.CorporateGroup) error {
	const q = `INSERT INTO corporate_groups (id, name, purchase_id, manager_id, teacher_id, teacher_join_code, student_link_id, semesters, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		g.ID, g.Name, g.PurchaseID, g.ManagerID, g.TeacherID,
		g.TeacherJoinCode, g.StudentLinkID, g.Semesters, g.CreatedAt,
	)
	return err
}

// GetByID fetches a single corporate group by its ID.
func (r *CorporateGroupRepo) GetByID(ctx context.Context, id string) (*subscription.CorporateGroup, error) {
	const q = `SELECT id, name, purchase_id, manager_id, teacher_id, teacher_join_code, student_link_id, semesters, created_at
		FROM corporate_groups WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	g, err := scanCorporateGroupRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	return g, err
}

// GetByJoinCode fetches a corporate group by its teacher join code.
func (r *CorporateGroupRepo) GetByJoinCode(ctx context.Context, code string) (*subscription.CorporateGroup, error) {
	const q = `SELECT id, name, purchase_id, manager_id, teacher_id, teacher_join_code, student_link_id, semesters, created_at
		FROM corporate_groups WHERE teacher_join_code = ?`
	row := r.db.QueryRowContext(ctx, q, code)
	g, err := scanCorporateGroupRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	return g, err
}

// GetByManagerID returns all corporate groups managed by a given manager.
func (r *CorporateGroupRepo) GetByManagerID(ctx context.Context, managerID int64) ([]*subscription.CorporateGroup, error) {
	const q = `SELECT id, name, purchase_id, manager_id, teacher_id, teacher_join_code, student_link_id, semesters, created_at
		FROM corporate_groups WHERE manager_id = ? ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, q, managerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*subscription.CorporateGroup
	for rows.Next() {
		g, err := scanCorporateGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// GetByTeacherID returns all corporate groups claimed by a given teacher.
func (r *CorporateGroupRepo) GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.CorporateGroup, error) {
	const q = `SELECT id, name, purchase_id, manager_id, teacher_id, teacher_join_code, student_link_id, semesters, created_at
		FROM corporate_groups WHERE teacher_id = ? ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, q, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*subscription.CorporateGroup
	for rows.Next() {
		g, err := scanCorporateGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// GetByStudentLinkID fetches a corporate group by its student invite link ID.
func (r *CorporateGroupRepo) GetByStudentLinkID(ctx context.Context, linkID string) (*subscription.CorporateGroup, error) {
	const q = `SELECT id, name, purchase_id, manager_id, teacher_id, teacher_join_code, student_link_id, semesters, created_at
		FROM corporate_groups WHERE student_link_id = ?`
	row := r.db.QueryRowContext(ctx, q, linkID)
	g, err := scanCorporateGroupRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	return g, err
}

// ClaimByTeacher assigns a teacher to a group atomically.
// Returns ErrGroupAlreadyClaimed if the group already has a teacher.
func (r *CorporateGroupRepo) ClaimByTeacher(ctx context.Context, groupID string, teacherID int64) error {
	// First check if already claimed.
	var existingTeacherID sql.NullInt64
	err := r.db.QueryRowContext(ctx, `SELECT teacher_id FROM corporate_groups WHERE id = ?`, groupID).Scan(&existingTeacherID)
	if err == sql.ErrNoRows {
		return apperrors.ErrNotFound
	}
	if err != nil {
		return err
	}
	if existingTeacherID.Valid {
		return apperrors.ErrGroupAlreadyClaimed
	}

	const q = `UPDATE corporate_groups SET teacher_id = ? WHERE id = ? AND teacher_id IS NULL`
	result, err := r.db.ExecContext(ctx, q, teacherID, groupID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// Race condition: another goroutine claimed it.
		return apperrors.ErrGroupAlreadyClaimed
	}
	return nil
}

// UpdateName updates the display name of a corporate group.
func (r *CorporateGroupRepo) UpdateName(ctx context.Context, groupID, name string) error {
	const q = `UPDATE corporate_groups SET name = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, name, groupID)
	return err
}

// GetAllPurchases returns every corporate purchase across all managers, newest first.
func (r *CorporateGroupRepo) GetAllPurchases(ctx context.Context) ([]*subscription.CorporatePurchase, error) {
	const q = `SELECT payment_id, manager_id, groups_count, semesters, total_amount, created_at
	           FROM corporate_purchases ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var purchases []*subscription.CorporatePurchase
	for rows.Next() {
		p := &subscription.CorporatePurchase{}
		if err := rows.Scan(&p.PaymentID, &p.ManagerID, &p.GroupsCount, &p.Semesters, &p.TotalAmount, &p.CreatedAt); err != nil {
			return nil, err
		}
		purchases = append(purchases, p)
	}
	return purchases, rows.Err()
}

// GetAllGroups returns every corporate group across all purchases, ordered by created_at.
func (r *CorporateGroupRepo) GetAllGroups(ctx context.Context) ([]*subscription.CorporateGroup, error) {
	const q = `SELECT id, name, purchase_id, manager_id, teacher_id,
	           teacher_join_code, student_link_id, semesters, created_at
	           FROM corporate_groups ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []*subscription.CorporateGroup
	for rows.Next() {
		g, err := scanCorporateGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// scanCorporateGroup scans a row from a *sql.Rows result.
func scanCorporateGroup(rows *sql.Rows) (*subscription.CorporateGroup, error) {
	g := &subscription.CorporateGroup{}
	var teacherID sql.NullInt64
	err := rows.Scan(&g.ID, &g.Name, &g.PurchaseID, &g.ManagerID, &teacherID,
		&g.TeacherJoinCode, &g.StudentLinkID, &g.Semesters, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	if teacherID.Valid {
		g.TeacherID = &teacherID.Int64
	}
	return g, nil
}

// scanCorporateGroupRow scans a single *sql.Row result.
func scanCorporateGroupRow(row *sql.Row) (*subscription.CorporateGroup, error) {
	g := &subscription.CorporateGroup{}
	var teacherID sql.NullInt64
	err := row.Scan(&g.ID, &g.Name, &g.PurchaseID, &g.ManagerID, &teacherID,
		&g.TeacherJoinCode, &g.StudentLinkID, &g.Semesters, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	if teacherID.Valid {
		g.TeacherID = &teacherID.Int64
	}
	return g, nil
}
