// Package admin provides use cases for administrative content and user management.
package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
)

// AnalyticsResult holds aggregated system metrics.
type AnalyticsResult struct {
	TotalUsers          int `json:"total_users"`
	ActiveSubscriptions int `json:"active_subscriptions"`
	ActivePromoCodes    int `json:"active_promo_codes"`
	TotalModules        int `json:"total_modules"`
	TotalTestAttempts   int `json:"total_test_attempts"`
}

// UseCase orchestrates admin operations.
type UseCase struct {
	modules    interfaces.ModuleRepository
	themes     interfaces.ThemeRepository
	mnemonics  interfaces.MnemonicRepository
	tests      interfaces.TestRepository
	promoCodes interfaces.PromoCodeRepository
	users      interfaces.UserRepository
	db         *sql.DB
}

// NewUseCase creates a new admin UseCase.
func NewUseCase(
	modules interfaces.ModuleRepository,
	themes interfaces.ThemeRepository,
	mnemonics interfaces.MnemonicRepository,
	tests interfaces.TestRepository,
	promoCodes interfaces.PromoCodeRepository,
	users interfaces.UserRepository,
	db *sql.DB,
) *UseCase {
	return &UseCase{
		modules:    modules,
		themes:     themes,
		mnemonics:  mnemonics,
		tests:      tests,
		promoCodes: promoCodes,
		users:      users,
		db:         db,
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

	if orderNum == 0 {
		maxNum, err := uc.modules.GetMaxOrderNum(ctx)
		if err != nil {
			return nil, err
		}
		orderNum = maxNum + 1
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

// DeleteModule deletes a module by ID.
func (uc *UseCase) DeleteModule(ctx context.Context, id int) error {
	return uc.modules.Delete(ctx, id)
}

// CreateTheme creates a new theme within a module.
func (uc *UseCase) CreateTheme(ctx context.Context, moduleID int, name, desc string, orderNum int, isIntro, isLocked bool, estimatedMins *int) (*content.Theme, error) {
	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}

	if orderNum == 0 {
		maxNum, err := uc.themes.GetMaxOrderNum(ctx, moduleID)
		if err != nil {
			return nil, err
		}
		orderNum = maxNum + 1
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

// UpdateTheme updates an existing theme's editable fields.
func (uc *UseCase) UpdateTheme(ctx context.Context, id int, name string, desc *string, orderNum int, isLocked bool, estimatedMins *int) (*content.Theme, error) {
	t, err := uc.themes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	t.Name = name
	t.Description = desc
	t.OrderNum = orderNum
	t.IsLocked = isLocked
	t.EstimatedTimeMinutes = estimatedMins

	return uc.themes.Update(ctx, t)
}

// DeleteTheme deletes a theme by ID.
func (uc *UseCase) DeleteTheme(ctx context.Context, id int) error {
	return uc.themes.Delete(ctx, id)
}

// CreateMnemonic creates a new mnemonic for a theme.
func (uc *UseCase) CreateMnemonic(ctx context.Context, themeID int, typ content.MnemonicType, text, s3Key *string, orderNum int) (*content.Mnemonic, error) {
	if orderNum == 0 {
		maxNum, err := uc.mnemonics.GetMaxOrderNum(ctx, themeID)
		if err != nil {
			return nil, err
		}
		orderNum = maxNum + 1
	}

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

// UpdateMnemonic updates an existing mnemonic's editable fields.
func (uc *UseCase) UpdateMnemonic(ctx context.Context, id int, contentText *string, s3Key *string, orderNum int) (*content.Mnemonic, error) {
	mn := &content.Mnemonic{
		ID:          id,
		ContentText: contentText,
		S3ImageKey:  s3Key,
		OrderNum:    orderNum,
	}
	return uc.mnemonics.Update(ctx, mn)
}

// DeleteMnemonic deletes a mnemonic by ID.
func (uc *UseCase) DeleteMnemonic(ctx context.Context, id int) error {
	return uc.mnemonics.Delete(ctx, id)
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

// UpdateTest updates an existing test's editable fields.
func (uc *UseCase) UpdateTest(ctx context.Context, id int, difficulty, passingScore int, shuffleQ, shuffleA bool, questions []content.Question) (*content.Test, error) {
	t, err := uc.tests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	t.Difficulty = difficulty
	t.PassingScore = passingScore
	t.ShuffleQuestions = shuffleQ
	t.ShuffleAnswers = shuffleA
	t.Questions = questions

	return uc.tests.Update(ctx, t)
}

// DeleteTest deletes a test by ID.
func (uc *UseCase) DeleteTest(ctx context.Context, id int) error {
	return uc.tests.Delete(ctx, id)
}

// GetUsers returns a paginated, optionally filtered list of users and the total count.
func (uc *UseCase) GetUsers(ctx context.Context, role *user.Role, subStatus *user.SubscriptionStatus, limit, offset int) ([]*user.User, int, error) {
	roleStr := ""
	if role != nil {
		roleStr = string(*role)
	}
	subStatusStr := ""
	if subStatus != nil {
		subStatusStr = string(*subStatus)
	}
	return uc.users.GetAll(ctx, roleStr, subStatusStr, limit, offset)
}

// GetModules returns all modules ordered by order_num.
func (uc *UseCase) GetModules(ctx context.Context) ([]*content.Module, error) {
	return uc.modules.GetAll(ctx)
}

// GetModuleByID returns a single module by ID.
func (uc *UseCase) GetModuleByID(ctx context.Context, id int) (*content.Module, error) {
	return uc.modules.GetByID(ctx, id)
}

// GetThemeByID returns a single theme by ID.
func (uc *UseCase) GetThemeByID(ctx context.Context, id int) (*content.Theme, error) {
	return uc.themes.GetByID(ctx, id)
}

// GetTestByID returns a single test by ID.
func (uc *UseCase) GetTestByID(ctx context.Context, id int) (*content.Test, error) {
	return uc.tests.GetByID(ctx, id)
}

// GetAllThemes returns all themes ordered by module and order_num.
func (uc *UseCase) GetAllThemes(ctx context.Context) ([]*content.Theme, error) {
	rows, err := uc.db.QueryContext(ctx, `
		SELECT id, module_id, name, description, order_num, is_introduction, is_locked,
		       estimated_time_minutes, created_at
		FROM themes ORDER BY module_id, order_num`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var themes []*content.Theme
	for rows.Next() {
		var t content.Theme
		var isIntroInt, isLockedInt int
		if err := rows.Scan(&t.ID, &t.ModuleID, &t.Name, &t.Description, &t.OrderNum,
			&isIntroInt, &isLockedInt, &t.EstimatedTimeMinutes, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.IsIntroduction = isIntroInt != 0
		t.IsLocked = isLockedInt != 0
		themes = append(themes, &t)
	}
	return themes, rows.Err()
}

// GetMnemonicByID returns a single mnemonic by ID.
func (uc *UseCase) GetMnemonicByID(ctx context.Context, id int) (*content.Mnemonic, error) {
	row := uc.db.QueryRowContext(ctx,
		`SELECT id, theme_id, type, content_text, s3_image_key, order_num, created_at FROM mnemonics WHERE id = ?`, id)
	var m content.Mnemonic
	var typeStr string
	if err := row.Scan(&m.ID, &m.ThemeID, &typeStr, &m.ContentText, &m.S3ImageKey, &m.OrderNum, &m.CreatedAt); err != nil {
		return nil, err
	}
	m.Type = content.MnemonicType(typeStr)
	return &m, nil
}

// GetAllMnemonics returns all mnemonics ordered by theme and order_num.
func (uc *UseCase) GetAllMnemonics(ctx context.Context) ([]*content.Mnemonic, error) {
	rows, err := uc.db.QueryContext(ctx,
		`SELECT id, theme_id, type, content_text, s3_image_key, order_num, created_at
		 FROM mnemonics ORDER BY theme_id, order_num`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*content.Mnemonic
	for rows.Next() {
		var m content.Mnemonic
		var typeStr string
		if err := rows.Scan(&m.ID, &m.ThemeID, &typeStr, &m.ContentText, &m.S3ImageKey, &m.OrderNum, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Type = content.MnemonicType(typeStr)
		list = append(list, &m)
	}
	return list, rows.Err()
}

// GetAllTests returns all tests ordered by theme_id.
func (uc *UseCase) GetAllTests(ctx context.Context) ([]*content.Test, error) {
	rows, err := uc.db.QueryContext(ctx,
		`SELECT id, theme_id, questions_json, difficulty, passing_score, shuffle_questions, shuffle_answers, created_at
		 FROM tests ORDER BY theme_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*content.Test
	for rows.Next() {
		var t content.Test
		var qJSON string
		var shuffleQInt, shuffleAInt int
		if err := rows.Scan(&t.ID, &t.ThemeID, &qJSON, &t.Difficulty, &t.PassingScore,
			&shuffleQInt, &shuffleAInt, &t.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(qJSON), &t.Questions)
		t.ShuffleQuestions = shuffleQInt != 0
		t.ShuffleAnswers = shuffleAInt != 0
		list = append(list, &t)
	}
	return list, rows.Err()
}

// GetAllPromoCodes returns all promo codes ordered by created_at desc.
func (uc *UseCase) GetAllPromoCodes(ctx context.Context) ([]*subscription.PromoCode, error) {
	rows, err := uc.db.QueryContext(ctx,
		`SELECT code, university_name, teacher_id, max_activations, remaining,
		        status, expires_at, created_by_admin_id, activated_at, created_at
		 FROM promo_codes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*subscription.PromoCode
	for rows.Next() {
		var p subscription.PromoCode
		var statusStr string
		if err := rows.Scan(&p.Code, &p.UniversityName, &p.TeacherID, &p.MaxActivations, &p.Remaining,
			&statusStr, &p.ExpiresAt, &p.CreatedByAdminID, &p.ActivatedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Status = subscription.PromoCodeStatus(statusStr)
		list = append(list, &p)
	}
	return list, rows.Err()
}

// CreateUser creates a new user by Telegram ID.
func (uc *UseCase) CreateUser(ctx context.Context, telegramID int64, role user.Role, subStatus user.SubscriptionStatus) (*user.User, error) {
	u := &user.User{
		TelegramID:           telegramID,
		Role:                 role,
		SubscriptionStatus:   subStatus,
		Language:             "ru",
		Timezone:             "UTC",
		NotificationsEnabled: true,
	}
	if err := uc.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return uc.users.GetByID(ctx, telegramID)
}

// GetUser returns a single user by Telegram ID.
func (uc *UseCase) GetUser(ctx context.Context, telegramID int64) (*user.User, error) {
	return uc.users.GetByID(ctx, telegramID)
}

// UpdateUserState updates role and subscription status of a user.
func (uc *UseCase) UpdateUserState(ctx context.Context, telegramID int64, role *user.Role, subStatus *user.SubscriptionStatus) (*user.User, error) {
	u, err := uc.users.GetByID(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	if role != nil {
		u.Role = *role
	}
	if subStatus != nil {
		u.SubscriptionStatus = *subStatus
	}
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetAnalytics returns aggregated system metrics.
func (uc *UseCase) GetAnalytics(ctx context.Context) (*AnalyticsResult, error) {
	var result AnalyticsResult

	queries := []struct {
		dest  *int
		query string
	}{
		{&result.TotalUsers, "SELECT COUNT(*) FROM users"},
		{&result.ActiveSubscriptions, "SELECT COUNT(*) FROM subscriptions WHERE status = 'active'"},
		{&result.ActivePromoCodes, "SELECT COUNT(*) FROM promo_codes WHERE status = 'active'"},
		{&result.TotalModules, "SELECT COUNT(*) FROM modules"},
		{&result.TotalTestAttempts, "SELECT COUNT(*) FROM test_attempts"},
	}

	for _, q := range queries {
		if err := uc.db.QueryRowContext(ctx, q.query).Scan(q.dest); err != nil {
			return nil, err
		}
	}

	return &result, nil
}
