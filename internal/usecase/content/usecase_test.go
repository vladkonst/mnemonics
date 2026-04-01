package content_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	ucContent "github.com/vladkonst/mnemonics/internal/usecase/content"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── Hand-written mocks ────────────────────────────────────────────────────────

type mockModuleRepo struct {
	modules []*content.Module
}

func (m *mockModuleRepo) GetAll(ctx context.Context) ([]*content.Module, error) {
	return m.modules, nil
}
func (m *mockModuleRepo) GetByID(ctx context.Context, id int) (*content.Module, error) {
	for _, mod := range m.modules {
		if mod.ID == id {
			return mod, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockModuleRepo) Create(ctx context.Context, mod *content.Module) error   { return nil }
func (m *mockModuleRepo) Update(ctx context.Context, mod *content.Module) error   { return nil }
func (m *mockModuleRepo) Delete(ctx context.Context, id int) error                { return nil }
func (m *mockModuleRepo) GetMaxOrderNum(ctx context.Context) (int, error)         { return 0, nil }

type mockThemeRepo struct {
	themes        map[int]*content.Theme
	byModule      map[int][]*content.Theme
	previousTheme map[int]*content.Theme // themeID → previous theme (nil or absent if first)
}

func (m *mockThemeRepo) GetByModuleID(ctx context.Context, moduleID int) ([]*content.Theme, error) {
	return m.byModule[moduleID], nil
}
func (m *mockThemeRepo) GetByID(ctx context.Context, id int) (*content.Theme, error) {
	if t, ok := m.themes[id]; ok {
		return t, nil
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockThemeRepo) Create(ctx context.Context, t *content.Theme) error { return nil }
func (m *mockThemeRepo) Update(ctx context.Context, t *content.Theme) (*content.Theme, error) {
	return t, nil
}
func (m *mockThemeRepo) Delete(ctx context.Context, id int) error                           { return nil }
func (m *mockThemeRepo) GetMaxOrderNum(ctx context.Context, moduleID int) (int, error)      { return 0, nil }
func (m *mockThemeRepo) GetPreviousTheme(ctx context.Context, themeID int) (*content.Theme, error) {
	prev, ok := m.previousTheme[themeID]
	if !ok || prev == nil {
		return nil, apperrors.ErrNotFound
	}
	return prev, nil
}

type mockMnemonicRepo struct{}

func (m *mockMnemonicRepo) GetByThemeID(ctx context.Context, themeID int) ([]*content.Mnemonic, error) {
	return nil, nil
}
func (m *mockMnemonicRepo) Create(ctx context.Context, mn *content.Mnemonic) error { return nil }
func (m *mockMnemonicRepo) Update(ctx context.Context, mn *content.Mnemonic) (*content.Mnemonic, error) {
	return mn, nil
}
func (m *mockMnemonicRepo) Delete(ctx context.Context, id int) error                              { return nil }
func (m *mockMnemonicRepo) GetMaxOrderNum(ctx context.Context, themeID int) (int, error)          { return 0, nil }

type mockTestRepo struct {
	tests map[int]*content.Test // themeID → test
}

func (m *mockTestRepo) GetByThemeID(ctx context.Context, themeID int) (*content.Test, error) {
	if t, ok := m.tests[themeID]; ok {
		return t, nil
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockTestRepo) GetByID(ctx context.Context, id int) (*content.Test, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockTestRepo) Create(ctx context.Context, t *content.Test) error { return nil }
func (m *mockTestRepo) Update(ctx context.Context, t *content.Test) (*content.Test, error) {
	return t, nil
}
func (m *mockTestRepo) Delete(ctx context.Context, id int) error { return nil }

type mockProgressRepo struct {
	data map[string]*progress.UserProgress
}

func progressKey(userID int64, themeID int) string {
	return fmt.Sprintf("%d-%d", userID, themeID)
}

func (m *mockProgressRepo) Upsert(ctx context.Context, p *progress.UserProgress) error {
	m.data[progressKey(p.UserID, p.ThemeID)] = p
	return nil
}
func (m *mockProgressRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) (*progress.UserProgress, error) {
	p, ok := m.data[progressKey(userID, themeID)]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return p, nil
}
func (m *mockProgressRepo) GetByUser(ctx context.Context, userID int64) ([]*progress.UserProgress, error) {
	var result []*progress.UserProgress
	for _, p := range m.data {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}
func (m *mockProgressRepo) GetByUserAndModule(ctx context.Context, userID int64, moduleID int) ([]*progress.UserProgress, error) {
	return nil, nil
}
func (m *mockProgressRepo) CountCompletedByUser(ctx context.Context, userID int64) (int, error) {
	count := 0
	for _, p := range m.data {
		if p.UserID == userID && p.IsCompleted() {
			count++
		}
	}
	return count, nil
}

type mockAttemptRepo struct {
	attempts map[string]*progress.TestAttempt
}

func (m *mockAttemptRepo) Create(ctx context.Context, a *progress.TestAttempt) error {
	m.attempts[a.AttemptID] = a
	return nil
}
func (m *mockAttemptRepo) GetByAttemptID(ctx context.Context, attemptID string) (*progress.TestAttempt, error) {
	a, ok := m.attempts[attemptID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return a, nil
}
func (m *mockAttemptRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) ([]*progress.TestAttempt, error) {
	var result []*progress.TestAttempt
	for _, a := range m.attempts {
		if a.UserID == userID && a.ThemeID == themeID {
			result = append(result, a)
		}
	}
	return result, nil
}

type mockSubscriptionRepo struct {
	active    map[int64]*subscription.Subscription
	byPayment map[string]*subscription.Subscription
}

func (m *mockSubscriptionRepo) Create(ctx context.Context, s *subscription.Subscription) error {
	m.byPayment[s.PaymentID] = s
	m.active[s.UserID] = s
	return nil
}
func (m *mockSubscriptionRepo) GetActiveByUserID(ctx context.Context, userID int64) (*subscription.Subscription, error) {
	s, ok := m.active[userID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return s, nil
}
func (m *mockSubscriptionRepo) GetByPaymentID(ctx context.Context, paymentID string) (*subscription.Subscription, error) {
	s, ok := m.byPayment[paymentID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return s, nil
}

type mockStorageService struct{}

func (m *mockStorageService) UploadFile(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	return nil
}
func (m *mockStorageService) PresignURL(ctx context.Context, s3Key string) (string, error) {
	return "https://cdn.example.com/" + s3Key, nil
}

// ── Constructor helpers ───────────────────────────────────────────────────────

func newTestUseCase(
	themeRepo *mockThemeRepo,
	progressRepo *mockProgressRepo,
	subRepo *mockSubscriptionRepo,
) *ucContent.UseCase {
	return ucContent.NewUseCase(
		&mockModuleRepo{},
		themeRepo,
		&mockMnemonicRepo{},
		&mockTestRepo{tests: map[int]*content.Test{}},
		progressRepo,
		&mockAttemptRepo{attempts: map[string]*progress.TestAttempt{}},
		subRepo,
		&mockStorageService{},
	)
}

func newFullUseCase(
	themeRepo *mockThemeRepo,
	testRepo *mockTestRepo,
	progressRepo *mockProgressRepo,
	attemptRepo *mockAttemptRepo,
	subRepo *mockSubscriptionRepo,
) *ucContent.UseCase {
	return ucContent.NewUseCase(
		&mockModuleRepo{},
		themeRepo,
		&mockMnemonicRepo{},
		testRepo,
		progressRepo,
		attemptRepo,
		subRepo,
		&mockStorageService{},
	)
}

// ── Tests: CheckThemeAccess ───────────────────────────────────────────────────

func TestCheckThemeAccess_WithActiveSubscription(t *testing.T) {
	userID := int64(1001)
	themeID := 5

	subRepo := &mockSubscriptionRepo{
		active: map[int64]*subscription.Subscription{
			userID: {
				PaymentID: "pay-1",
				UserID:    userID,
				Status:    subscription.SubscriptionPlanStatusActive,
				Type:      subscription.SubscriptionTypePersonal,
			},
		},
		byPayment: map[string]*subscription.Subscription{},
	}

	uc := newTestUseCase(
		&mockThemeRepo{
			themes:        map[int]*content.Theme{},
			byModule:      map[int][]*content.Theme{},
			previousTheme: map[int]*content.Theme{},
		},
		&mockProgressRepo{data: map[string]*progress.UserProgress{}},
		subRepo,
	)

	result, err := uc.CheckThemeAccess(context.Background(), userID, themeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Accessible {
		t.Error("expected Accessible=true for user with active subscription")
	}
	if result.AccessType != "subscription" {
		t.Errorf("expected AccessType=subscription, got %s", result.AccessType)
	}
}

func TestCheckThemeAccess_NoSubscription_PrevCompleted(t *testing.T) {
	userID := int64(1002)
	themeID := 10
	prevThemeID := 9

	prevTheme := &content.Theme{ID: prevThemeID, ModuleID: 1, Name: "Previous Theme", OrderNum: 1}
	currTheme := &content.Theme{ID: themeID, ModuleID: 1, Name: "Current Theme", OrderNum: 2}

	themeRepo := &mockThemeRepo{
		themes: map[int]*content.Theme{
			themeID:     currTheme,
			prevThemeID: prevTheme,
		},
		byModule: map[int][]*content.Theme{},
		previousTheme: map[int]*content.Theme{
			themeID: prevTheme,
		},
	}

	now := time.Now().UTC()
	score := 80
	progressRepo := &mockProgressRepo{
		data: map[string]*progress.UserProgress{
			fmt.Sprintf("%d-%d", userID, prevThemeID): {
				UserID:      userID,
				ThemeID:     prevThemeID,
				Status:      progress.StatusCompleted,
				Score:       &score,
				CompletedAt: &now,
			},
		},
	}

	subRepo := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}

	uc := newTestUseCase(themeRepo, progressRepo, subRepo)

	result, err := uc.CheckThemeAccess(context.Background(), userID, themeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Accessible {
		t.Error("expected Accessible=true when previous theme is completed")
	}
	if result.AccessType != "sequential" {
		t.Errorf("expected AccessType=sequential, got %s", result.AccessType)
	}
}

func TestCheckThemeAccess_NoSubscription_PrevNotCompleted(t *testing.T) {
	userID := int64(1003)
	themeID := 10
	prevThemeID := 9

	prevTheme := &content.Theme{ID: prevThemeID, ModuleID: 1, Name: "Previous Theme", OrderNum: 1}
	currTheme := &content.Theme{ID: themeID, ModuleID: 1, Name: "Current Theme", OrderNum: 2}

	themeRepo := &mockThemeRepo{
		themes: map[int]*content.Theme{
			themeID:     currTheme,
			prevThemeID: prevTheme,
		},
		byModule: map[int][]*content.Theme{},
		previousTheme: map[int]*content.Theme{
			themeID: prevTheme,
		},
	}

	// Previous theme started but NOT completed.
	progressRepo := &mockProgressRepo{
		data: map[string]*progress.UserProgress{
			fmt.Sprintf("%d-%d", userID, prevThemeID): {
				UserID:  userID,
				ThemeID: prevThemeID,
				Status:  progress.StatusStarted,
			},
		},
	}

	subRepo := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}

	uc := newTestUseCase(themeRepo, progressRepo, subRepo)

	result, err := uc.CheckThemeAccess(context.Background(), userID, themeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Accessible {
		t.Error("expected Accessible=false when previous theme is not completed")
	}
	if result.RequiredThemeID == nil || *result.RequiredThemeID != prevThemeID {
		t.Errorf("expected RequiredThemeID=%d", prevThemeID)
	}
}

func TestCheckThemeAccess_FirstTheme_AlwaysAccessible(t *testing.T) {
	userID := int64(1004)
	themeID := 1

	themeRepo := &mockThemeRepo{
		themes: map[int]*content.Theme{
			themeID: {ID: themeID, ModuleID: 1, Name: "Intro", OrderNum: 1, IsIntroduction: true},
		},
		byModule:      map[int][]*content.Theme{},
		previousTheme: map[int]*content.Theme{}, // no previous → ErrNotFound
	}

	subRepo := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}

	uc := newTestUseCase(themeRepo, &mockProgressRepo{data: map[string]*progress.UserProgress{}}, subRepo)

	result, err := uc.CheckThemeAccess(context.Background(), userID, themeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Accessible {
		t.Error("expected Accessible=true for first theme with no previous")
	}
}

// ── Tests: SubmitTestAttempt ──────────────────────────────────────────────────

func TestSubmitTestAttempt_Passed(t *testing.T) {
	userID := int64(2001)
	themeID := 20
	attemptID := "attempt-uuid-pass"

	themeRepo := &mockThemeRepo{
		themes: map[int]*content.Theme{
			themeID: {ID: themeID, ModuleID: 1, Name: "Theme", OrderNum: 1},
		},
		byModule:      map[int][]*content.Theme{1: {{ID: themeID, ModuleID: 1, Name: "Theme", OrderNum: 1}}},
		previousTheme: map[int]*content.Theme{},
	}

	testObj := &content.Test{
		ID:           100,
		ThemeID:      themeID,
		PassingScore: 60,
		Difficulty:   2,
		Questions: []content.Question{
			{ID: 1, Text: "Q1", Type: content.QuestionTypeMultipleChoice, CorrectAnswer: "A", Options: []string{"A", "B"}},
			{ID: 2, Text: "Q2", Type: content.QuestionTypeMultipleChoice, CorrectAnswer: "B", Options: []string{"A", "B"}},
		},
	}
	testRepo := &mockTestRepo{tests: map[int]*content.Test{themeID: testObj}}

	attempt := &progress.TestAttempt{
		ID:        1,
		UserID:    userID,
		ThemeID:   themeID,
		TestID:    testObj.ID,
		AttemptID: attemptID,
		StartedAt: time.Now().UTC(),
	}
	attemptRepo := &mockAttemptRepo{attempts: map[string]*progress.TestAttempt{attemptID: attempt}}
	progressRepo := &mockProgressRepo{data: map[string]*progress.UserProgress{}}
	subRepo := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}

	uc := newFullUseCase(themeRepo, testRepo, progressRepo, attemptRepo, subRepo)

	answers := []progress.AnswerItem{
		{QuestionID: 1, Answer: "A"}, // correct
		{QuestionID: 2, Answer: "B"}, // correct
	}

	result, err := uc.SubmitTestAttempt(context.Background(), userID, attemptID, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected Passed=true when all answers are correct")
	}
	if result.Score != 100 {
		t.Errorf("expected Score=100, got %d", result.Score)
	}
	if result.CorrectAnswers != 2 {
		t.Errorf("expected CorrectAnswers=2, got %d", result.CorrectAnswers)
	}

	// Verify progress was updated to completed.
	p, err := progressRepo.GetByUserAndTheme(context.Background(), userID, themeID)
	if err != nil {
		t.Fatalf("expected progress to be saved: %v", err)
	}
	if !p.IsCompleted() {
		t.Errorf("expected progress status=completed, got %s", p.Status)
	}
}

func TestSubmitTestAttempt_Failed(t *testing.T) {
	userID := int64(2002)
	themeID := 21
	attemptID := "attempt-uuid-fail"

	themeRepo := &mockThemeRepo{
		themes: map[int]*content.Theme{
			themeID: {ID: themeID, ModuleID: 1, Name: "Theme 2", OrderNum: 2},
		},
		byModule: map[int][]*content.Theme{1: {
			{ID: 20, ModuleID: 1, Name: "Theme 1", OrderNum: 1},
			{ID: themeID, ModuleID: 1, Name: "Theme 2", OrderNum: 2},
		}},
		previousTheme: map[int]*content.Theme{},
	}

	testObj := &content.Test{
		ID:           101,
		ThemeID:      themeID,
		PassingScore: 80,
		Difficulty:   2,
		Questions: []content.Question{
			{ID: 1, Text: "Q1", CorrectAnswer: "A", Options: []string{"A", "B"}},
			{ID: 2, Text: "Q2", CorrectAnswer: "B", Options: []string{"A", "B"}},
		},
	}
	testRepo := &mockTestRepo{tests: map[int]*content.Test{themeID: testObj}}

	attempt := &progress.TestAttempt{
		ID:        2,
		UserID:    userID,
		ThemeID:   themeID,
		TestID:    testObj.ID,
		AttemptID: attemptID,
		StartedAt: time.Now().UTC(),
	}
	attemptRepo := &mockAttemptRepo{attempts: map[string]*progress.TestAttempt{attemptID: attempt}}
	progressRepo := &mockProgressRepo{data: map[string]*progress.UserProgress{}}
	subRepo := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}

	uc := newFullUseCase(themeRepo, testRepo, progressRepo, attemptRepo, subRepo)

	// Only 1 of 2 correct → 50% < 80% passing score.
	answers := []progress.AnswerItem{
		{QuestionID: 1, Answer: "A"}, // correct
		{QuestionID: 2, Answer: "A"}, // wrong
	}

	result, err := uc.SubmitTestAttempt(context.Background(), userID, attemptID, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected Passed=false when score < passing score")
	}
	if result.Score != 50 {
		t.Errorf("expected Score=50, got %d", result.Score)
	}

	p, err := progressRepo.GetByUserAndTheme(context.Background(), userID, themeID)
	if err != nil {
		t.Fatalf("expected progress to be saved: %v", err)
	}
	if p.Status != progress.StatusFailed {
		t.Errorf("expected Status=failed, got %s", p.Status)
	}
}

func TestSubmitTestAttempt_IdempotentResubmit(t *testing.T) {
	userID := int64(2003)
	themeID := 22
	attemptID := "attempt-uuid-idempotent"

	themeRepo := &mockThemeRepo{
		themes: map[int]*content.Theme{
			themeID: {ID: themeID, ModuleID: 1, Name: "Idempotent Theme", OrderNum: 1},
		},
		byModule:      map[int][]*content.Theme{},
		previousTheme: map[int]*content.Theme{},
	}

	testObj := &content.Test{
		ID:           102,
		ThemeID:      themeID,
		PassingScore: 60,
		Difficulty:   1,
		Questions: []content.Question{
			{ID: 1, Text: "Q1", CorrectAnswer: "A", Options: []string{"A", "B"}},
		},
	}
	testRepo := &mockTestRepo{tests: map[int]*content.Test{themeID: testObj}}

	submittedAt := time.Now().UTC()
	// Pre-submitted attempt.
	attempt := &progress.TestAttempt{
		ID:          3,
		UserID:      userID,
		ThemeID:     themeID,
		TestID:      testObj.ID,
		AttemptID:   attemptID,
		Score:       100,
		Passed:      true,
		StartedAt:   submittedAt.Add(-30 * time.Second),
		SubmittedAt: &submittedAt,
	}
	attemptRepo := &mockAttemptRepo{attempts: map[string]*progress.TestAttempt{attemptID: attempt}}
	progressRepo := &mockProgressRepo{data: map[string]*progress.UserProgress{}}
	subRepo := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}

	uc := newFullUseCase(themeRepo, testRepo, progressRepo, attemptRepo, subRepo)

	// Submit again — should return cached result without re-grading.
	result, err := uc.SubmitTestAttempt(context.Background(), userID, attemptID, nil)
	if err != nil {
		t.Fatalf("unexpected error on idempotent re-submit: %v", err)
	}
	if !result.Passed {
		t.Error("expected cached Passed=true on re-submit")
	}
	if result.Score != 100 {
		t.Errorf("expected cached Score=100, got %d", result.Score)
	}
}
