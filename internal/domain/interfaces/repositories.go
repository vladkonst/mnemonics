// Package interfaces defines repository and service contracts for the domain layer.
// All implementations live in the infrastructure/repository layers.
// The domain layer has zero external dependencies.
package interfaces

import (
	"context"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
)

// ── User ─────────────────────────────────────────────────────────────────────

// UserRepository abstracts persistence for the User aggregate.
type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	GetByID(ctx context.Context, telegramID int64) (*user.User, error)
	Update(ctx context.Context, u *user.User) error
	Exists(ctx context.Context, telegramID int64) (bool, error)
}

// ── Content ──────────────────────────────────────────────────────────────────

// ModuleRepository abstracts persistence for Module aggregates.
type ModuleRepository interface {
	GetAll(ctx context.Context) ([]*content.Module, error)
	GetByID(ctx context.Context, id int) (*content.Module, error)
	Create(ctx context.Context, m *content.Module) error
	Update(ctx context.Context, m *content.Module) error
}

// ThemeRepository abstracts persistence for Theme entities.
type ThemeRepository interface {
	GetByModuleID(ctx context.Context, moduleID int) ([]*content.Theme, error)
	GetByID(ctx context.Context, id int) (*content.Theme, error)
	Create(ctx context.Context, t *content.Theme) error
	// GetPreviousTheme returns the theme with order_num = theme.order_num - 1 in the same module.
	GetPreviousTheme(ctx context.Context, themeID int) (*content.Theme, error)
}

// MnemonicRepository abstracts persistence for Mnemonic entities.
type MnemonicRepository interface {
	GetByThemeID(ctx context.Context, themeID int) ([]*content.Mnemonic, error)
	Create(ctx context.Context, m *content.Mnemonic) error
}

// TestRepository abstracts persistence for Test aggregates.
type TestRepository interface {
	GetByThemeID(ctx context.Context, themeID int) (*content.Test, error)
	GetByID(ctx context.Context, id int) (*content.Test, error)
	Create(ctx context.Context, t *content.Test) error
}

// ── Progress ─────────────────────────────────────────────────────────────────

// ProgressRepository abstracts persistence for UserProgress aggregates.
type ProgressRepository interface {
	// Upsert creates or updates progress for (userID, themeID).
	Upsert(ctx context.Context, p *progress.UserProgress) error
	GetByUserAndTheme(ctx context.Context, userID int64, themeID int) (*progress.UserProgress, error)
	GetByUser(ctx context.Context, userID int64) ([]*progress.UserProgress, error)
	GetByUserAndModule(ctx context.Context, userID int64, moduleID int) ([]*progress.UserProgress, error)
	// CountCompletedByUser returns the number of completed themes for a user.
	CountCompletedByUser(ctx context.Context, userID int64) (int, error)
}

// TestAttemptRepository abstracts persistence for TestAttempt records.
type TestAttemptRepository interface {
	Create(ctx context.Context, a *progress.TestAttempt) error
	GetByAttemptID(ctx context.Context, attemptID string) (*progress.TestAttempt, error)
	GetByUserAndTheme(ctx context.Context, userID int64, themeID int) ([]*progress.TestAttempt, error)
}

// ── Subscription ─────────────────────────────────────────────────────────────

// PromoCodeRepository abstracts persistence for PromoCode aggregates.
type PromoCodeRepository interface {
	GetByCode(ctx context.Context, code string) (*subscription.PromoCode, error)
	Update(ctx context.Context, p *subscription.PromoCode) error
	Create(ctx context.Context, p *subscription.PromoCode) error
	Deactivate(ctx context.Context, code string) error
	GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.PromoCode, error)
}

// SubscriptionRepository abstracts persistence for Subscription records.
type SubscriptionRepository interface {
	Create(ctx context.Context, s *subscription.Subscription) error
	GetActiveByUserID(ctx context.Context, userID int64) (*subscription.Subscription, error)
	GetByPaymentID(ctx context.Context, paymentID string) (*subscription.Subscription, error)
}

// TeacherStudentRepository abstracts the teacher↔student relationship table.
type TeacherStudentRepository interface {
	AddStudent(ctx context.Context, teacherID, studentID int64, promoCode string) error
	GetStudentsByTeacher(ctx context.Context, teacherID int64) ([]*user.User, error)
	IsTeacherStudent(ctx context.Context, teacherID, studentID int64) (bool, error)
}
