package sqlite

import (
	"context"
	"database/sql"

	"github.com/vladkonst/mnemonics/internal/domain/user"
)

// TeacherStudentRepo implements interfaces.TeacherStudentRepository using SQLite.
type TeacherStudentRepo struct {
	db *sql.DB
}

func NewTeacherStudentRepo(db *sql.DB) *TeacherStudentRepo {
	return &TeacherStudentRepo{db: db}
}

func (r *TeacherStudentRepo) AddStudent(ctx context.Context, teacherID, studentID int64, promoCode string) error {
	const q = `
		INSERT OR IGNORE INTO teacher_promo_students (teacher_id, student_id, promo_code)
		VALUES (?, ?, ?)`

	_, err := r.db.ExecContext(ctx, q, teacherID, studentID, promoCode)
	return err
}

func (r *TeacherStudentRepo) GetStudentsByTeacher(ctx context.Context, teacherID int64) ([]*user.User, error) {
	const q = `
		SELECT u.telegram_id, u.role, u.subscription_status, u.university_code,
		       u.pending_payment_id, u.username,
		       u.language, u.timezone, u.notifications_enabled, u.last_activity_at, u.created_at
		FROM users u
		JOIN teacher_promo_students ts ON ts.student_id = u.telegram_id
		WHERE ts.teacher_id = ?`

	rows, err := r.db.QueryContext(ctx, q, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u, err := scanUserRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *TeacherStudentRepo) IsTeacherStudent(ctx context.Context, teacherID, studentID int64) (bool, error) {
	const q = `
		SELECT COUNT(1) FROM teacher_promo_students
		WHERE teacher_id = ? AND student_id = ?`

	var count int
	err := r.db.QueryRowContext(ctx, q, teacherID, studentID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func scanUserRows(rows *sql.Rows) (*user.User, error) {
	var u user.User
	var roleStr, statusStr string
	var notifInt int

	err := rows.Scan(
		&u.TelegramID,
		&roleStr,
		&statusStr,
		&u.UniversityCode,
		&u.PendingPaymentID,
		&u.Username,
		&u.Language,
		&u.Timezone,
		&notifInt,
		&u.LastActivityAt,
		&u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = user.Role(roleStr)
	u.SubscriptionStatus = user.SubscriptionStatus(statusStr)
	u.NotificationsEnabled = notifInt != 0
	return &u, nil
}
