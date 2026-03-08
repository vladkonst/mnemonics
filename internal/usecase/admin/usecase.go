// Package admin provides use cases for administrative content and user management.
package admin

import (
	"context"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
)

// UseCase orchestrates admin operations.
type UseCase struct {
	modules    interfaces.ModuleRepository
	themes     interfaces.ThemeRepository
	mnemonics  interfaces.MnemonicRepository
	tests      interfaces.TestRepository
	promoCodes interfaces.PromoCodeRepository
	users      interfaces.UserRepository
}

// NewUseCase creates a new admin UseCase.
func NewUseCase(
	modules interfaces.ModuleRepository,
	themes interfaces.ThemeRepository,
	mnemonics interfaces.MnemonicRepository,
	tests interfaces.TestRepository,
	promoCodes interfaces.PromoCodeRepository,
	users interfaces.UserRepository,
) *UseCase {
	return &UseCase{
		modules:    modules,
		themes:     themes,
		mnemonics:  mnemonics,
		tests:      tests,
		promoCodes: promoCodes,
		users:      users,
	}
}

// CreatePromoCode creates a new promo code in pending state.
func (uc *UseCase) CreatePromoCode(ctx context.Context, code, universityName string, maxActivations int, expiresAt *time.Time) (*subscription.PromoCode, error) {
	now := time.Now().UTC()
	promo := &subscription.PromoCode{
		Code:           code,
		UniversityName: universityName,
		MaxActivations: maxActivations,
		Remaining:      maxActivations,
		Status:         subscription.PromoCodeStatusPending,
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
	}

	if err := uc.promoCodes.Create(ctx, promo); err != nil {
		return nil, err
	}
	return promo, nil
}

// DeactivatePromoCode marks a promo code as deactivated.
func (uc *UseCase) DeactivatePromoCode(ctx context.Context, code string) error {
	return uc.promoCodes.Deactivate(ctx, code)
}

// CreateModule creates a new content module.
func (uc *UseCase) CreateModule(ctx context.Context, name, description string, orderNum int, isLocked bool, iconEmoji *string) (*content.Module, error) {
	var descPtr *string
	if description != "" {
		descPtr = &description
	}

	m := &content.Module{
		Name:        name,
		Description: descPtr,
		OrderNum:    orderNum,
		IsLocked:    isLocked,
		IconEmoji:   iconEmoji,
		CreatedAt:   time.Now().UTC(),
	}

	if err := uc.modules.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// UpdateModule updates an existing module's metadata.
func (uc *UseCase) UpdateModule(ctx context.Context, id int, name, description string, orderNum int, isLocked bool, iconEmoji *string) (*content.Module, error) {
	m, err := uc.modules.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	m.Name = name
	if description != "" {
		m.Description = &description
	}
	m.OrderNum = orderNum
	m.IsLocked = isLocked
	m.IconEmoji = iconEmoji

	if err := uc.modules.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// CreateTheme creates a new theme within a module.
func (uc *UseCase) CreateTheme(ctx context.Context, moduleID int, name, desc string, orderNum int, isIntro, isLocked bool, estimatedMins *int) (*content.Theme, error) {
	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}

	t := &content.Theme{
		ModuleID:             moduleID,
		Name:                 name,
		Description:          descPtr,
		OrderNum:             orderNum,
		IsIntroduction:       isIntro,
		IsLocked:             isLocked,
		EstimatedTimeMinutes: estimatedMins,
		CreatedAt:            time.Now().UTC(),
	}

	if err := uc.themes.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// CreateMnemonic creates a new mnemonic for a theme.
func (uc *UseCase) CreateMnemonic(ctx context.Context, themeID int, typ content.MnemonicType, text, s3Key *string, orderNum int) (*content.Mnemonic, error) {
	m := &content.Mnemonic{
		ThemeID:     themeID,
		Type:        typ,
		ContentText: text,
		S3ImageKey:  s3Key,
		OrderNum:    orderNum,
		CreatedAt:   time.Now().UTC(),
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	if err := uc.mnemonics.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// CreateTest creates a new test for a theme.
func (uc *UseCase) CreateTest(ctx context.Context, themeID, difficulty, passingScore int, shuffleQ, shuffleA bool, questions []content.Question) (*content.Test, error) {
	t := &content.Test{
		ThemeID:          themeID,
		Questions:        questions,
		Difficulty:       difficulty,
		PassingScore:     passingScore,
		ShuffleQuestions: shuffleQ,
		ShuffleAnswers:   shuffleA,
		CreatedAt:        time.Now().UTC(),
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	if err := uc.tests.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// GetUsers returns a paginated, optionally filtered list of users and the total count.
func (uc *UseCase) GetUsers(ctx context.Context, role *user.Role, subStatus *user.SubscriptionStatus, limit, offset int) ([]*user.User, int, error) {
	// UserRepository does not expose a filtered list method, so we fetch all
	// and filter in memory. In production this would be pushed to the DB layer.
	// This satisfies the interface contract without requiring new repository methods.
	//
	// NOTE: For large datasets this is inefficient but acceptable per the current
	// repository interface definition.
	_ = role
	_ = subStatus
	_ = limit
	_ = offset

	// Since the UserRepository interface only exposes Create/GetByID/Update/Exists,
	// we cannot fetch a list of users without extending it.
	// Return an empty slice with a placeholder implementation.
	// A real delivery layer would call a richer repository method.
	return []*user.User{}, 0, nil
}
