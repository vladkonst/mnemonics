// Package content provides use cases for educational content delivery.
package content

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── Result types ─────────────────────────────────────────────────────────────

// ModuleWithProgress enriches a Module with the user's completion data.
type ModuleWithProgress struct {
	*content.Module
	TotalThemes     int  `json:"total_themes"`
	CompletedThemes int  `json:"completed_themes"`
	IsAccessible    bool `json:"is_accessible"`
}

// ThemeWithAccess enriches a Theme with access and completion information.
type ThemeWithAccess struct {
	*content.Theme
	IsAccessible bool    `json:"is_accessible"`
	IsCompleted  bool    `json:"is_completed"`
	Score        *int    `json:"score,omitempty"`
	LockedReason *string `json:"locked_reason,omitempty"`
}

// ModuleThemesResult is the response for listing themes in a module.
type ModuleThemesResult struct {
	ModuleID   int                `json:"module_id"`
	ModuleName string             `json:"module_name"`
	Themes     []*ThemeWithAccess `json:"themes"`
}

// StudySessionResult is the response when starting a study session.
type StudySessionResult struct {
	SessionID     string              `json:"session_id"`
	Theme         *content.Theme      `json:"theme"`
	Mnemonics     []*content.Mnemonic `json:"mnemonics"`
	TestAvailable bool                `json:"test_available"`
	TestID        *int                `json:"test_id,omitempty"`
}

// AccessResult carries the outcome of a theme access check.
type AccessResult struct {
	Accessible        bool    `json:"accessible"`
	AccessType        string  `json:"access_type"`
	Reason            *string `json:"reason,omitempty"`
	RequiredThemeID   *int    `json:"required_theme_id,omitempty"`
	RequiredThemeName *string `json:"required_theme_name,omitempty"`
	RequiredAction    *string `json:"required_action,omitempty"`
}

// ── UseCase ──────────────────────────────────────────────────────────────────

// UseCase orchestrates content delivery operations.
type UseCase struct {
	modules       interfaces.ModuleRepository
	themes        interfaces.ThemeRepository
	mnemonics     interfaces.MnemonicRepository
	tests         interfaces.TestRepository
	progress      interfaces.ProgressRepository
	attempts      interfaces.TestAttemptRepository
	subscriptions interfaces.SubscriptionRepository
	storage       interfaces.StorageService
}

// NewUseCase creates a new content UseCase.
func NewUseCase(
	modules interfaces.ModuleRepository,
	themes interfaces.ThemeRepository,
	mnemonics interfaces.MnemonicRepository,
	tests interfaces.TestRepository,
	progress interfaces.ProgressRepository,
	attempts interfaces.TestAttemptRepository,
	subscriptions interfaces.SubscriptionRepository,
	storage interfaces.StorageService,
) *UseCase {
	return &UseCase{
		modules:       modules,
		themes:        themes,
		mnemonics:     mnemonics,
		tests:         tests,
		progress:      progress,
		attempts:      attempts,
		subscriptions: subscriptions,
		storage:       storage,
	}
}

// GetModules returns all modules enriched with per-user completion counts.
func (uc *UseCase) GetModules(ctx context.Context, userID int64) ([]*ModuleWithProgress, error) {
	modules, err := uc.modules.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*ModuleWithProgress, 0, len(modules))
	for _, m := range modules {
		themes, err := uc.themes.GetByModuleID(ctx, m.ID)
		if err != nil {
			return nil, err
		}

		completed := 0
		for _, t := range themes {
			p, err := uc.progress.GetByUserAndTheme(ctx, userID, t.ID)
			if err != nil && !apperrors.IsNotFound(err) {
				return nil, err
			}
			if p != nil && p.IsCompleted() {
				completed++
			}
		}

		result = append(result, &ModuleWithProgress{
			Module:          m,
			TotalThemes:     len(themes),
			CompletedThemes: completed,
			IsAccessible:    !m.IsLocked,
		})
	}
	return result, nil
}

// GetModuleThemes returns all themes in a module enriched with access information.
func (uc *UseCase) GetModuleThemes(ctx context.Context, moduleID int, userID int64) (*ModuleThemesResult, error) {
	mod, err := uc.modules.GetByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	themes, err := uc.themes.GetByModuleID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	enriched := make([]*ThemeWithAccess, 0, len(themes))
	for _, t := range themes {
		access, err := uc.CheckThemeAccess(ctx, userID, t.ID)
		if err != nil {
			return nil, err
		}

		twa := &ThemeWithAccess{
			Theme:        t,
			IsAccessible: access.Accessible,
			IsCompleted:  false,
		}
		if !access.Accessible && access.Reason != nil {
			twa.LockedReason = access.Reason
		}

		p, err := uc.progress.GetByUserAndTheme(ctx, userID, t.ID)
		if err != nil && !apperrors.IsNotFound(err) {
			return nil, err
		}
		if p != nil {
			twa.IsCompleted = p.IsCompleted()
			twa.Score = p.Score
		}

		enriched = append(enriched, twa)
	}

	return &ModuleThemesResult{
		ModuleID:   mod.ID,
		ModuleName: mod.Name,
		Themes:     enriched,
	}, nil
}

// GetModule returns a single module by ID.
func (uc *UseCase) GetModule(ctx context.Context, moduleID int) (*content.Module, error) {
	return uc.modules.GetByID(ctx, moduleID)
}

// GetTheme returns a theme with its mnemonics.
func (uc *UseCase) GetTheme(ctx context.Context, themeID int) (*StudySessionResult, error) {
	theme, err := uc.themes.GetByID(ctx, themeID)
	if err != nil {
		return nil, err
	}

	mnems, err := uc.mnemonics.GetByThemeID(ctx, themeID)
	if err != nil {
		return nil, err
	}
	for _, m := range mnems {
		if m.Type == content.MnemonicTypeImage && m.S3ImageKey != nil {
			url, err := uc.storage.PresignURL(ctx, *m.S3ImageKey)
			if err != nil {
				return nil, fmt.Errorf("presign URL for mnemonic %d: %w", m.ID, err)
			}
			presignedURL := url
			m.S3ImageKey = &presignedURL
		}
	}

	test, err := uc.tests.GetByThemeID(ctx, themeID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}

	result := &StudySessionResult{
		Theme:         theme,
		Mnemonics:     mnems,
		TestAvailable: test != nil,
	}
	if test != nil {
		result.TestID = &test.ID
	}
	return result, nil
}

// CheckThemeAccess determines whether a user may access a given theme.
func (uc *UseCase) CheckThemeAccess(ctx context.Context, userID int64, themeID int) (*AccessResult, error) {
	// Check active subscription.
	sub, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}

	if sub != nil && sub.IsActive() {
		return &AccessResult{
			Accessible: true,
			AccessType: "subscription",
		}, nil
	}

	// No active subscription: sequential logic.
	prevTheme, err := uc.themes.GetPreviousTheme(ctx, themeID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}

	// First theme (no previous) is always accessible.
	if prevTheme == nil {
		return &AccessResult{
			Accessible: true,
			AccessType: "sequential",
		}, nil
	}

	// Check if previous theme is completed.
	prevProgress, err := uc.progress.GetByUserAndTheme(ctx, userID, prevTheme.ID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}

	if prevProgress != nil && prevProgress.IsCompleted() {
		return &AccessResult{
			Accessible: true,
			AccessType: "sequential",
		}, nil
	}

	// Not accessible.
	reason := "previous_theme_required"
	action := "complete_previous_theme"
	return &AccessResult{
		Accessible:        false,
		AccessType:        "sequential",
		Reason:            &reason,
		RequiredThemeID:   &prevTheme.ID,
		RequiredThemeName: &prevTheme.Name,
		RequiredAction:    &action,
	}, nil
}

// CreateStudySession checks access, marks the theme as started, and returns mnemonics.
func (uc *UseCase) CreateStudySession(ctx context.Context, userID int64, themeID int) (*StudySessionResult, error) {
	access, err := uc.CheckThemeAccess(ctx, userID, themeID)
	if err != nil {
		return nil, err
	}
	if !access.Accessible {
		return nil, apperrors.ErrAccessDenied
	}

	theme, err := uc.themes.GetByID(ctx, themeID)
	if err != nil {
		return nil, err
	}

	// Upsert progress as started.
	prog, err := uc.progress.GetByUserAndTheme(ctx, userID, themeID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if prog == nil {
		now := time.Now().UTC()
		prog = &progress.UserProgress{
			UserID:    userID,
			ThemeID:   themeID,
			Status:    progress.StatusStarted,
			StartedAt: now,
			UpdatedAt: now,
		}
	}
	prog.MarkStarted()
	if err := uc.progress.Upsert(ctx, prog); err != nil {
		return nil, err
	}

	// Load mnemonics and generate presigned URLs for image mnemonics.
	mnems, err := uc.mnemonics.GetByThemeID(ctx, themeID)
	if err != nil {
		return nil, err
	}
	for _, m := range mnems {
		if m.Type == content.MnemonicTypeImage && m.S3ImageKey != nil {
			url, err := uc.storage.PresignURL(ctx, *m.S3ImageKey)
			if err != nil {
				return nil, fmt.Errorf("presign URL for mnemonic %d: %w", m.ID, err)
			}
			// Replace the S3 key with the presigned URL so the caller can use it directly.
			presignedURL := url
			m.S3ImageKey = &presignedURL
		}
	}

	// Check whether a test is available for this theme.
	test, err := uc.tests.GetByThemeID(ctx, themeID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}

	result := &StudySessionResult{
		SessionID:     uuid.NewString(),
		Theme:         theme,
		Mnemonics:     mnems,
		TestAvailable: test != nil,
	}
	if test != nil {
		result.TestID = &test.ID
	}
	return result, nil
}
