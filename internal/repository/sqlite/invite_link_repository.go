package sqlite

import (
	"context"
	"database/sql"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// InviteLinkRepo implements interfaces.InviteLinkRepository using SQLite.
type InviteLinkRepo struct {
	db *sql.DB
}

func NewInviteLinkRepo(db *sql.DB) *InviteLinkRepo {
	return &InviteLinkRepo{db: db}
}

func (r *InviteLinkRepo) Create(ctx context.Context, link *subscription.InviteLink) error {
	const q = `INSERT INTO invite_links (id, teacher_id, max_activations, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, link.ID, link.TeacherID, link.MaxActivations, link.CreatedAt)
	return err
}

func (r *InviteLinkRepo) GetByID(ctx context.Context, id string) (*subscription.InviteLink, error) {
	const q = `SELECT id, teacher_id, max_activations, created_at FROM invite_links WHERE id = ?`
	link := &subscription.InviteLink{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(&link.ID, &link.TeacherID, &link.MaxActivations, &link.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	return link, err
}

func (r *InviteLinkRepo) GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.InviteLink, error) {
	const q = `SELECT id, teacher_id, max_activations, created_at FROM invite_links WHERE teacher_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*subscription.InviteLink
	for rows.Next() {
		l := &subscription.InviteLink{}
		if err := rows.Scan(&l.ID, &l.TeacherID, &l.MaxActivations, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (r *InviteLinkRepo) GetAll(ctx context.Context, limit, offset int) ([]*subscription.InviteLink, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invite_links`).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `SELECT id, teacher_id, max_activations, created_at FROM invite_links ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var links []*subscription.InviteLink
	for rows.Next() {
		l := &subscription.InviteLink{}
		if err := rows.Scan(&l.ID, &l.TeacherID, &l.MaxActivations, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		links = append(links, l)
	}
	return links, total, rows.Err()
}

func (r *InviteLinkRepo) AddActivation(ctx context.Context, a *subscription.InviteLinkActivation) error {
	const q = `INSERT OR IGNORE INTO invite_link_activations (invite_link_id, user_id, activated_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, a.InviteLinkID, a.UserID, a.ActivatedAt)
	return err
}

func (r *InviteLinkRepo) GetActivations(ctx context.Context, linkID string) ([]*subscription.InviteLinkActivation, error) {
	const q = `SELECT id, invite_link_id, user_id, activated_at FROM invite_link_activations WHERE invite_link_id = ? ORDER BY activated_at DESC`
	rows, err := r.db.QueryContext(ctx, q, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var acts []*subscription.InviteLinkActivation
	for rows.Next() {
		a := &subscription.InviteLinkActivation{}
		if err := rows.Scan(&a.ID, &a.InviteLinkID, &a.UserID, &a.ActivatedAt); err != nil {
			return nil, err
		}
		acts = append(acts, a)
	}
	return acts, rows.Err()
}

func (r *InviteLinkRepo) CountActivations(ctx context.Context, linkID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invite_link_activations WHERE invite_link_id = ?`, linkID).Scan(&n)
	return n, err
}
