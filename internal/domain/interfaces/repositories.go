// Package interfaces defines repository and service contracts for the domain layer.
// All implementations live in the infrastructure/repository layers.
// The domain layer has zero external dependencies.
package interfaces

import (
	"context"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/feedback"
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
	Delete(ctx context.Context, telegramID int64) error
	Exists(ctx context.Context, telegramID int64) (bool, error)
	GetAll(ctx context.Context, role, subStatus string, limit, offset int) ([]*user.User, int, error)
}

// ── Content ──────────────────────────────────────────────────────────────────

// ModuleRepository abstracts persistence for Module aggregates.
type ModuleRepository interface {
	GetAll(ctx context.Context) ([]*content.Module, error)
	GetByID(ctx context.Context, id int) (*content.Module, error)
	Create(ctx context.Context, m *content.Module) error
	Update(ctx context.Context, m *content.Module) error
	Delete(ctx context.Context, id int) error
	GetMaxOrderNum(ctx context.Context) (int, error)
}

// ThemeRepository abstracts persistence for Theme entities.
type ThemeRepository interface {
	GetByModuleID(ctx context.Context, moduleID int) ([]*content.Theme, error)
	GetByID(ctx context.Context, id int) (*content.Theme, error)
	Create(ctx context.Context, t *content.Theme) error
	// GetPreviousTheme returns the theme with order_num = theme.order_num - 1 in the same module.
	GetPreviousTheme(ctx context.Context, themeID int) (*content.Theme, error)
	Update(ctx context.Context, theme *content.Theme) (*content.Theme, error)
	Delete(ctx context.Context, id int) error
	GetMaxOrderNum(ctx context.Context, moduleID int) (int, error)
}

// MnemonicRepository abstracts persistence for Mnemonic entities.
type MnemonicRepository interface {
	GetByThemeID(ctx context.Context, themeID int) ([]*content.Mnemonic, error)
	Create(ctx context.Context, m *content.Mnemonic) error
	Update(ctx context.Context, m *content.Mnemonic) (*content.Mnemonic, error)
	Delete(ctx context.Context, id int) error
	GetMaxOrderNum(ctx context.Context, themeID int) (int, error)
}

// TestRepository abstracts persistence for Test aggregates.
type TestRepository interface {
	GetByThemeID(ctx context.Context, themeID int) (*content.Test, error)
	GetByID(ctx context.Context, id int) (*content.Test, error)
	Create(ctx context.Context, t *content.Test) error
	Update(ctx context.Context, t *content.Test) (*content.Test, error)
	Delete(ctx context.Context, id int) error
}

// ModuleTestRepository abstracts persistence for ModuleTest aggregates.
type ModuleTestRepository interface {
	GetAll(ctx context.Context) ([]*content.ModuleTest, error)
	GetByModuleID(ctx context.Context, moduleID int) ([]*content.ModuleTest, error)
	GetByID(ctx context.Context, id int) (*content.ModuleTest, error)
	Create(ctx context.Context, t *content.ModuleTest) error
	Update(ctx context.Context, t *content.ModuleTest) (*content.ModuleTest, error)
	Delete(ctx context.Context, id int) error
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

// ModuleTestAttemptRepository abstracts persistence for dynamically-generated module test attempts.
type ModuleTestAttemptRepository interface {
	Create(ctx context.Context, a *progress.ModuleTestAttempt) error
	Update(ctx context.Context, a *progress.ModuleTestAttempt) error
	GetByAttemptID(ctx context.Context, attemptID string) (*progress.ModuleTestAttempt, error)
	GetByUserAndModule(ctx context.Context, userID int64, moduleID int) ([]*progress.ModuleTestAttempt, error)
}

// ── Subscription ─────────────────────────────────────────────────────────────

// SubscriptionRepository abstracts persistence for Subscription records.
type SubscriptionRepository interface {
	Create(ctx context.Context, s *subscription.Subscription) error
	GetActiveByUserID(ctx context.Context, userID int64) (*subscription.Subscription, error)
	GetByPaymentID(ctx context.Context, paymentID string) (*subscription.Subscription, error)
}

// TeacherStudentRepository abstracts the teacher↔student relationship table.
type TeacherStudentRepository interface {
	AddStudent(ctx context.Context, teacherID, studentID int64, joinRef string) error
	GetStudentsByTeacher(ctx context.Context, teacherID int64) ([]*user.User, error)
	IsTeacherStudent(ctx context.Context, teacherID, studentID int64) (bool, error)
}

// ── Invite Links ─────────────────────────────────────────────────────────────

// InviteLinkRepository abstracts persistence for teacher-generated invite links.
type InviteLinkRepository interface {
	Create(ctx context.Context, link *subscription.InviteLink) error
	GetByID(ctx context.Context, id string) (*subscription.InviteLink, error)
	GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.InviteLink, error)
	GetAll(ctx context.Context, limit, offset int) ([]*subscription.InviteLink, int, error)
	AddActivation(ctx context.Context, a *subscription.InviteLinkActivation) error
	GetActivations(ctx context.Context, linkID string) ([]*subscription.InviteLinkActivation, error)
	CountActivations(ctx context.Context, linkID string) (int, error)
}

// ── Corporate Groups ──────────────────────────────────────────────────────────

// CorporateGroupRepository abstracts persistence for corporate purchase and group entities.
type CorporateGroupRepository interface {
	CreatePurchase(ctx context.Context, p *subscription.CorporatePurchase) error
	GetPurchaseByPaymentID(ctx context.Context, paymentID string) (*subscription.CorporatePurchase, error)
	GetPurchasesByManagerID(ctx context.Context, managerID int64) ([]*subscription.CorporatePurchase, error)
	GetAllPurchases(ctx context.Context) ([]*subscription.CorporatePurchase, error)
	GetGroupsByPurchaseID(ctx context.Context, purchaseID string) ([]*subscription.CorporateGroup, error)

	Create(ctx context.Context, g *subscription.CorporateGroup) error
	GetByID(ctx context.Context, id string) (*subscription.CorporateGroup, error)
	GetByJoinCode(ctx context.Context, code string) (*subscription.CorporateGroup, error)
	GetByManagerID(ctx context.Context, managerID int64) ([]*subscription.CorporateGroup, error)
	GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.CorporateGroup, error)
	GetByStudentLinkID(ctx context.Context, linkID string) (*subscription.CorporateGroup, error)
	GetAllGroups(ctx context.Context) ([]*subscription.CorporateGroup, error)
	ClaimByTeacher(ctx context.Context, groupID string, teacherID int64) error
	UpdateName(ctx context.Context, groupID, name string) error
}

// ── Feedback ─────────────────────────────────────────────────────────────────

// FeedbackRepository abstracts persistence for user feedback.
type FeedbackRepository interface {
	Create(ctx context.Context, f *feedback.Feedback) error
	GetAll(ctx context.Context, limit, offset int) ([]*feedback.Feedback, int, error)
	GetByID(ctx context.Context, id int) (*feedback.Feedback, error)
}
