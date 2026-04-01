package admin

import (
	"context"
	"testing"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── mocks ─────────────────────────────────────────────────────────────────────

type mockModuleRepo struct {
	data    map[int]*content.Module
	nextID  int
}

func newMockModuleRepo() *mockModuleRepo {
	return &mockModuleRepo{data: make(map[int]*content.Module), nextID: 1}
}

func (m *mockModuleRepo) GetAll(ctx context.Context) ([]*content.Module, error) { return nil, nil }
func (m *mockModuleRepo) GetByID(ctx context.Context, id int) (*content.Module, error) {
	mod, ok := m.data[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return mod, nil
}
func (m *mockModuleRepo) Create(ctx context.Context, mod *content.Module) error {
	mod.ID = m.nextID
	m.nextID++
	m.data[mod.ID] = mod
	return nil
}
func (m *mockModuleRepo) Update(ctx context.Context, mod *content.Module) error {
	m.data[mod.ID] = mod
	return nil
}
func (m *mockModuleRepo) Delete(ctx context.Context, id int) error {
	if _, ok := m.data[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.data, id)
	return nil
}
func (m *mockModuleRepo) GetMaxOrderNum(ctx context.Context) (int, error) {
	max := 0
	for _, mod := range m.data {
		if mod.OrderNum > max {
			max = mod.OrderNum
		}
	}
	return max, nil
}

type mockThemeRepo struct {
	data   map[int]*content.Theme
	nextID int
}

func newMockThemeRepo() *mockThemeRepo {
	return &mockThemeRepo{data: make(map[int]*content.Theme), nextID: 1}
}
func (m *mockThemeRepo) GetByModuleID(ctx context.Context, moduleID int) ([]*content.Theme, error) {
	return nil, nil
}
func (m *mockThemeRepo) GetByID(ctx context.Context, id int) (*content.Theme, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockThemeRepo) Create(ctx context.Context, t *content.Theme) error {
	t.ID = m.nextID
	m.nextID++
	m.data[t.ID] = t
	return nil
}
func (m *mockThemeRepo) GetPreviousTheme(ctx context.Context, themeID int) (*content.Theme, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockThemeRepo) Update(ctx context.Context, t *content.Theme) (*content.Theme, error) {
	m.data[t.ID] = t
	return t, nil
}
func (m *mockThemeRepo) Delete(ctx context.Context, id int) error {
	if _, ok := m.data[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.data, id)
	return nil
}
func (m *mockThemeRepo) GetMaxOrderNum(ctx context.Context, moduleID int) (int, error) {
	max := 0
	for _, t := range m.data {
		if t.ModuleID == moduleID && t.OrderNum > max {
			max = t.OrderNum
		}
	}
	return max, nil
}

type mockMnemonicRepo struct {
	data []*content.Mnemonic
}

func (m *mockMnemonicRepo) GetByThemeID(ctx context.Context, id int) ([]*content.Mnemonic, error) {
	return nil, nil
}
func (m *mockMnemonicRepo) Create(ctx context.Context, mn *content.Mnemonic) error {
	m.data = append(m.data, mn)
	return nil
}
func (m *mockMnemonicRepo) Update(ctx context.Context, mn *content.Mnemonic) (*content.Mnemonic, error) {
	return mn, nil
}
func (m *mockMnemonicRepo) Delete(ctx context.Context, id int) error { return nil }
func (m *mockMnemonicRepo) GetMaxOrderNum(ctx context.Context, themeID int) (int, error) {
	max := 0
	for _, mn := range m.data {
		if mn.ThemeID == themeID && mn.OrderNum > max {
			max = mn.OrderNum
		}
	}
	return max, nil
}

type mockTestRepo struct {
	data []*content.Test
}

func (m *mockTestRepo) GetByThemeID(ctx context.Context, id int) (*content.Test, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockTestRepo) GetByID(ctx context.Context, id int) (*content.Test, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockTestRepo) Create(ctx context.Context, t *content.Test) error {
	m.data = append(m.data, t)
	return nil
}
func (m *mockTestRepo) Update(ctx context.Context, t *content.Test) (*content.Test, error) {
	return t, nil
}
func (m *mockTestRepo) Delete(ctx context.Context, id int) error { return nil }

type mockPromoCodeRepo struct {
	data map[string]*subscription.PromoCode
}

func newMockPromoCodeRepo() *mockPromoCodeRepo {
	return &mockPromoCodeRepo{data: make(map[string]*subscription.PromoCode)}
}
func (m *mockPromoCodeRepo) GetByCode(ctx context.Context, code string) (*subscription.PromoCode, error) {
	p, ok := m.data[code]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return p, nil
}
func (m *mockPromoCodeRepo) Update(ctx context.Context, p *subscription.PromoCode) error {
	m.data[p.Code] = p
	return nil
}
func (m *mockPromoCodeRepo) Create(ctx context.Context, p *subscription.PromoCode) error {
	m.data[p.Code] = p
	return nil
}
func (m *mockPromoCodeRepo) Deactivate(ctx context.Context, code string) error {
	p, ok := m.data[code]
	if !ok {
		return apperrors.ErrNotFound
	}
	p.Status = subscription.PromoCodeStatusDeactivated
	return nil
}
func (m *mockPromoCodeRepo) GetByTeacherID(ctx context.Context, id int64) ([]*subscription.PromoCode, error) {
	return nil, nil
}
func (m *mockPromoCodeRepo) ConsumeOne(ctx context.Context, code string) error { return nil }

type mockUserRepo struct{}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error             { return nil }
func (m *mockUserRepo) Exists(ctx context.Context, id int64) (bool, error)         { return false, nil }
func (m *mockUserRepo) GetAll(ctx context.Context, role, subStatus string, limit, offset int) ([]*user.User, int, error) {
	return []*user.User{}, 0, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newUC() (*UseCase, *mockModuleRepo, *mockThemeRepo, *mockMnemonicRepo, *mockTestRepo, *mockPromoCodeRepo) {
	mods := newMockModuleRepo()
	themes := newMockThemeRepo()
	mnems := &mockMnemonicRepo{}
	tests := &mockTestRepo{}
	promos := newMockPromoCodeRepo()
	uc := NewUseCase(mods, themes, mnems, tests, promos, &mockUserRepo{}, nil)
	return uc, mods, themes, mnems, tests, promos
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestCreatePromoCode(t *testing.T) {
	uc, _, _, _, _, promos := newUC()

	p, err := uc.CreatePromoCode(context.Background(), "TEST2024", "TestU", 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Code != "TEST2024" {
		t.Errorf("Code = %q, want TEST2024", p.Code)
	}
	if p.Remaining != 10 {
		t.Errorf("Remaining = %d, want 10", p.Remaining)
	}
	if p.Status != subscription.PromoCodeStatusPending {
		t.Errorf("Status = %q, want pending", p.Status)
	}

	// With expiry
	future := time.Now().Add(24 * time.Hour)
	p2, err := uc.CreatePromoCode(context.Background(), "EXP2024", "U", 5, &future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p2.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}
	_ = promos
}

func TestDeactivatePromoCode(t *testing.T) {
	uc, _, _, _, _, promos := newUC()

	promos.data["DEL"] = &subscription.PromoCode{Code: "DEL", Status: subscription.PromoCodeStatusActive}
	err := uc.DeactivatePromoCode(context.Background(), "DEL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promos.data["DEL"].Status != subscription.PromoCodeStatusDeactivated {
		t.Error("promo code should be deactivated")
	}
}

func TestCreateModule(t *testing.T) {
	uc, mods, _, _, _, _ := newUC()

	m, err := uc.CreateModule(context.Background(), "Anatomy", "Study of the body", 1, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Anatomy" {
		t.Errorf("Name = %q, want Anatomy", m.Name)
	}
	if m.Description == nil || *m.Description != "Study of the body" {
		t.Error("Description should be set")
	}
	_ = mods
}

func TestCreateModule_EmptyDescription(t *testing.T) {
	uc, _, _, _, _, _ := newUC()

	m, err := uc.CreateModule(context.Background(), "Bio", "", 1, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Description != nil {
		t.Error("Description should be nil when empty")
	}
}

func TestUpdateModule_HappyPath(t *testing.T) {
	uc, mods, _, _, _, _ := newUC()

	_ , _ = uc.CreateModule(context.Background(), "Old", "Desc", 1, false, nil)
	modID := 1

	updated, err := uc.UpdateModule(context.Background(), modID, "New", "New Desc", 2, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "New" {
		t.Errorf("Name = %q, want New", updated.Name)
	}
	if !updated.IsLocked {
		t.Error("IsLocked should be true")
	}
	_ = mods
}

func TestUpdateModule_NotFound(t *testing.T) {
	uc, _, _, _, _, _ := newUC()

	_, err := uc.UpdateModule(context.Background(), 999, "X", "", 1, false, nil)
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestCreateTheme(t *testing.T) {
	uc, _, themes, _, _, _ := newUC()

	th, err := uc.CreateTheme(context.Background(), 1, "Bones", "Skeleton", 1, true, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th.Name != "Bones" {
		t.Errorf("Name = %q, want Bones", th.Name)
	}
	if !th.IsIntroduction {
		t.Error("IsIntroduction should be true")
	}
	_ = themes
}

func TestCreateMnemonic_Text(t *testing.T) {
	uc, _, _, mnems, _, _ := newUC()

	text := "Remember: bones are hard"
	m, err := uc.CreateMnemonic(context.Background(), 1, content.MnemonicTypeText, &text, nil, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Type != content.MnemonicTypeText {
		t.Errorf("Type = %q, want text", m.Type)
	}
	_ = mnems
}

func TestCreateMnemonic_Invalid(t *testing.T) {
	uc, _, _, _, _, _ := newUC()

	// text mnemonic with no text
	_, err := uc.CreateMnemonic(context.Background(), 1, content.MnemonicTypeText, nil, nil, 1)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestCreateTest_Valid(t *testing.T) {
	uc, _, _, _, tests, _ := newUC()

	questions := []content.Question{
		{ID: 1, Text: "Q?", Type: content.QuestionTypeMultipleChoice, CorrectAnswer: "A", OrderNum: 1},
	}
	test, err := uc.CreateTest(context.Background(), 1, 2, 70, false, false, questions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if test.PassingScore != 70 {
		t.Errorf("PassingScore = %d, want 70", test.PassingScore)
	}
	_ = tests
}

func TestCreateTest_Invalid(t *testing.T) {
	uc, _, _, _, _, _ := newUC()

	// No questions
	_, err := uc.CreateTest(context.Background(), 1, 2, 70, false, false, nil)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestGetUsers_ReturnsEmpty(t *testing.T) {
	uc, _, _, _, _, _ := newUC()

	users, total, err := uc.GetUsers(context.Background(), nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(users) != 0 {
		t.Errorf("len(users) = %d, want 0", len(users))
	}
}
