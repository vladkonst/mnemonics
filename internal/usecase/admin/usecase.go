// Package admin provides use cases for administrative content and user management.
package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/feedback"
	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// AnalyticsResult holds aggregated system metrics.
type AnalyticsResult struct {
	TotalUsers          int `json:"total_users"`
	ActiveSubscriptions int `json:"active_subscriptions"`
	ActiveInviteLinks   int `json:"active_invite_links"`
	TotalModules        int `json:"total_modules"`
	TotalTestAttempts   int `json:"total_test_attempts"`
}

// UseCase orchestrates admin operations.
type UseCase struct {
	modules         interfaces.ModuleRepository
	themes          interfaces.ThemeRepository
	mnemonics       interfaces.MnemonicRepository
	tests           interfaces.TestRepository
	moduleTests     interfaces.ModuleTestRepository
	users           interfaces.UserRepository
	teacherStudents interfaces.TeacherStudentRepository
	feedbackRepo    interfaces.FeedbackRepository
	inviteLinks     interfaces.InviteLinkRepository
	db              *sql.DB
}

// NewUseCase creates a new admin UseCase.
func NewUseCase(
	modules interfaces.ModuleRepository,
	themes interfaces.ThemeRepository,
	mnemonics interfaces.MnemonicRepository,
	tests interfaces.TestRepository,
	moduleTests interfaces.ModuleTestRepository,
	users interfaces.UserRepository,
	teacherStudents interfaces.TeacherStudentRepository,
	feedbackRepo interfaces.FeedbackRepository,
	inviteLinks interfaces.InviteLinkRepository,
	db *sql.DB,
) *UseCase {
	return &UseCase{
		modules:         modules,
		themes:          themes,
		mnemonics:       mnemonics,
		tests:           tests,
		moduleTests:     moduleTests,
		users:           users,
		teacherStudents: teacherStudents,
		feedbackRepo:    feedbackRepo,
		inviteLinks:     inviteLinks,
		db:              db,
	}
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
func (uc *UseCase) CreateMnemonic(ctx context.Context, themeID int, typ content.MnemonicType, text, s3Key, termRu, termLatin *string, orderNum int) (*content.Mnemonic, error) {
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
		TermRu:      termRu,
		TermLatin:   termLatin,
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
func (uc *UseCase) UpdateMnemonic(ctx context.Context, id int, contentText *string, s3Key, termRu, termLatin *string, orderNum int) (*content.Mnemonic, error) {
	mn := &content.Mnemonic{
		ID:          id,
		ContentText: contentText,
		S3ImageKey:  s3Key,
		TermRu:      termRu,
		TermLatin:   termLatin,
		OrderNum:    orderNum,
	}
	return uc.mnemonics.Update(ctx, mn)
}

// DeleteMnemonic deletes a mnemonic by ID.
func (uc *UseCase) DeleteMnemonic(ctx context.Context, id int) error {
	return uc.mnemonics.Delete(ctx, id)
}

// CreateTest creates a new test for a theme.
func (uc *UseCase) CreateTest(ctx context.Context, themeID, difficulty, passingScore int, shuffleQ bool, questions []content.Question) (*content.Test, error) {
	t := &content.Test{
		ThemeID:          themeID,
		Questions:        questions,
		Difficulty:       difficulty,
		PassingScore:     passingScore,
		ShuffleQuestions: shuffleQ,
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
func (uc *UseCase) UpdateTest(ctx context.Context, id int, difficulty, passingScore int, shuffleQ bool, questions []content.Question) (*content.Test, error) {
	t, err := uc.tests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	t.Difficulty = difficulty
	t.PassingScore = passingScore
	t.ShuffleQuestions = shuffleQ
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
		`SELECT id, theme_id, type, content_text, s3_image_key, term_ru, term_latin, order_num, created_at FROM mnemonics WHERE id = ?`, id)
	var m content.Mnemonic
	var typeStr string
	if err := row.Scan(&m.ID, &m.ThemeID, &typeStr, &m.ContentText, &m.S3ImageKey, &m.TermRu, &m.TermLatin, &m.OrderNum, &m.CreatedAt); err != nil {
		return nil, err
	}
	m.Type = content.MnemonicType(typeStr)
	return &m, nil
}

// GetAllMnemonics returns all mnemonics ordered by theme and order_num.
func (uc *UseCase) GetAllMnemonics(ctx context.Context) ([]*content.Mnemonic, error) {
	rows, err := uc.db.QueryContext(ctx,
		`SELECT id, theme_id, type, content_text, s3_image_key, term_ru, term_latin, order_num, created_at
		 FROM mnemonics ORDER BY theme_id, order_num`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*content.Mnemonic
	for rows.Next() {
		var m content.Mnemonic
		var typeStr string
		if err := rows.Scan(&m.ID, &m.ThemeID, &typeStr, &m.ContentText, &m.S3ImageKey, &m.TermRu, &m.TermLatin, &m.OrderNum, &m.CreatedAt); err != nil {
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
		`SELECT id, theme_id, questions_json, difficulty, passing_score, shuffle_questions, created_at
		 FROM tests ORDER BY theme_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*content.Test
	for rows.Next() {
		var t content.Test
		var qJSON string
		var shuffleQInt int
		if err := rows.Scan(&t.ID, &t.ThemeID, &qJSON, &t.Difficulty, &t.PassingScore,
			&shuffleQInt, &t.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(qJSON), &t.Questions)
		t.ShuffleQuestions = shuffleQInt != 0
		list = append(list, &t)
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

func (uc *UseCase) DeleteUser(ctx context.Context, telegramID int64) error {
	return uc.users.Delete(ctx, telegramID)
}

func (uc *UseCase) GetStudentsByTeacher(ctx context.Context, teacherID int64) ([]*user.User, error) {
	return uc.teacherStudents.GetStudentsByTeacher(ctx, teacherID)
}

// CreateModuleTest creates a module test with manually specified questions.
func (uc *UseCase) CreateModuleTest(ctx context.Context, moduleID int, name string, difficulty, passingScore int, shuffleQ bool, questions []content.Question) (*content.ModuleTest, error) {
	t := &content.ModuleTest{
		ModuleID:         moduleID,
		Name:             name,
		Questions:        questions,
		Difficulty:       difficulty,
		PassingScore:     passingScore,
		ShuffleQuestions: shuffleQ,
		CreatedAt:        time.Now().UTC(),
	}
	if err := t.Validate(); err != nil {
		return nil, apperrors.New("bad_request", err.Error(), apperrors.ErrInvalidInput)
	}
	if err := uc.moduleTests.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// GenerateModuleTest auto-generates a module test by sampling 50% of questions
// from each theme's test within the module. Passing score is set to 100%.
func (uc *UseCase) GenerateModuleTest(ctx context.Context, moduleID int) (*content.ModuleTest, error) {
	module, err := uc.modules.GetByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	themes, err := uc.themes.GetByModuleID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	var allQuestions []content.Question
	maxID := 0
	for _, th := range themes {
		test, err := uc.tests.GetByThemeID(ctx, th.ID)
		if err != nil {
			// Theme may have no test yet — skip.
			continue
		}
		pool := make([]content.Question, len(test.Questions))
		copy(pool, test.Questions)
		n := (len(pool)*50 + 99) / 100 // ceil of 50%
		if n > len(pool) {
			n = len(pool)
		}
		rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		allQuestions = append(allQuestions, pool[:n]...)
	}

	// Re-number question IDs sequentially within this test to avoid collisions.
	for i := range allQuestions {
		maxID++
		allQuestions[i].ID = maxID
	}

	t := &content.ModuleTest{
		ModuleID:         moduleID,
		Name:             "Тест модуля \"" + module.Name + "\"",
		Questions:        allQuestions,
		Difficulty:       1,
		PassingScore:     100,
		ShuffleQuestions: true,
		CreatedAt:        time.Now().UTC(),
	}
	if err := t.Validate(); err != nil {
		return nil, apperrors.New("bad_request", err.Error(), apperrors.ErrInvalidInput)
	}
	if err := uc.moduleTests.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// UpdateModuleTest updates an existing module test.
func (uc *UseCase) UpdateModuleTest(ctx context.Context, id int, name string, difficulty, passingScore int, shuffleQ bool, questions []content.Question) (*content.ModuleTest, error) {
	t, err := uc.moduleTests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Name = name
	t.Difficulty = difficulty
	t.PassingScore = passingScore
	t.ShuffleQuestions = shuffleQ
	t.Questions = questions
	return uc.moduleTests.Update(ctx, t)
}

// DeleteModuleTest deletes a module test by ID.
func (uc *UseCase) DeleteModuleTest(ctx context.Context, id int) error {
	return uc.moduleTests.Delete(ctx, id)
}

// GetModuleTestByID returns a single module test by ID.
func (uc *UseCase) GetModuleTestByID(ctx context.Context, id int) (*content.ModuleTest, error) {
	return uc.moduleTests.GetByID(ctx, id)
}

// GetModuleTestsByModule returns all module tests for a given module.
func (uc *UseCase) GetModuleTestsByModule(ctx context.Context, moduleID int) ([]*content.ModuleTest, error) {
	return uc.moduleTests.GetByModuleID(ctx, moduleID)
}

// GetAllModuleTests returns all module tests across all modules.
func (uc *UseCase) GetAllModuleTests(ctx context.Context) ([]*content.ModuleTest, error) {
	return uc.moduleTests.GetAll(ctx)
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
		{&result.ActiveInviteLinks, "SELECT COUNT(*) FROM invite_links"},
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

// GetFeedback returns a paginated list of all feedback entries.
func (uc *UseCase) GetFeedback(ctx context.Context, limit, offset int) ([]*feedback.Feedback, int, error) {
	return uc.feedbackRepo.GetAll(ctx, limit, offset)
}

// GetFeedbackByID returns a single feedback entry by ID.
func (uc *UseCase) GetFeedbackByID(ctx context.Context, id int) (*feedback.Feedback, error) {
	return uc.feedbackRepo.GetByID(ctx, id)
}

// GetInviteLinks returns a paginated list of all invite links.
func (uc *UseCase) GetInviteLinks(ctx context.Context, limit, offset int) ([]*subscription.InviteLink, int, error) {
	return uc.inviteLinks.GetAll(ctx, limit, offset)
}

// GetInviteLinkByID returns a single invite link with its activations.
func (uc *UseCase) GetInviteLinkByID(ctx context.Context, id string) (*subscription.InviteLinkWithStats, error) {
	link, err := uc.inviteLinks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	acts, err := uc.inviteLinks.GetActivations(ctx, id)
	if err != nil {
		return nil, err
	}
	return &subscription.InviteLinkWithStats{
		InviteLink:      *link,
		ActivationCount: len(acts),
		Activations:     acts,
	}, nil
}
