package progress

import (
	"context"
	"fmt"
	"testing"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── mocks ─────────────────────────────────────────────────────────────────────

func progKey(userID int64, themeID int) string {
	return fmt.Sprintf("%d-%d", userID, themeID)
}

type mockProgressRepo struct {
	data map[string]*progress.UserProgress
}

func newMockProgressRepo() *mockProgressRepo {
	return &mockProgressRepo{data: make(map[string]*progress.UserProgress)}
}

func (m *mockProgressRepo) Upsert(ctx context.Context, p *progress.UserProgress) error {
	m.data[progKey(p.UserID, p.ThemeID)] = p
	return nil
}

func (m *mockProgressRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) (*progress.UserProgress, error) {
	p, ok := m.data[progKey(userID, themeID)]
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

type mockAttemptRepo struct{}

func (m *mockAttemptRepo) Create(ctx context.Context, a *progress.TestAttempt) error { return nil }
func (m *mockAttemptRepo) GetByAttemptID(ctx context.Context, id string) (*progress.TestAttempt, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockAttemptRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) ([]*progress.TestAttempt, error) {
	return nil, nil
}

type mockTestRepo struct{}

func (m *mockTestRepo) GetByThemeID(ctx context.Context, themeID int) (*content.Test, error) {
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

type mockThemeRepo struct {
	data map[int]*content.Theme
}

func newMockThemeRepo(themes ...*content.Theme) *mockThemeRepo {
	m := &mockThemeRepo{data: make(map[int]*content.Theme)}
	for _, t := range themes {
		m.data[t.ID] = t
	}
	return m
}

func (m *mockThemeRepo) GetByModuleID(ctx context.Context, moduleID int) ([]*content.Theme, error) {
	var result []*content.Theme
	for _, t := range m.data {
		if t.ModuleID == moduleID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockThemeRepo) GetByID(ctx context.Context, id int) (*content.Theme, error) {
	t, ok := m.data[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return t, nil
}

func (m *mockThemeRepo) Create(ctx context.Context, t *content.Theme) error { return nil }
func (m *mockThemeRepo) Update(ctx context.Context, t *content.Theme) (*content.Theme, error) {
	return t, nil
}
func (m *mockThemeRepo) Delete(ctx context.Context, id int) error                      { return nil }
func (m *mockThemeRepo) GetMaxOrderNum(ctx context.Context, moduleID int) (int, error) { return 0, nil }

func (m *mockThemeRepo) GetPreviousTheme(ctx context.Context, themeID int) (*content.Theme, error) {
	return nil, apperrors.ErrNotFound
}

type mockModuleRepo struct {
	data map[int]*content.Module
}

func newMockModuleRepo(modules ...*content.Module) *mockModuleRepo {
	m := &mockModuleRepo{data: make(map[int]*content.Module)}
	for _, mod := range modules {
		m.data[mod.ID] = mod
	}
	return m
}

func (m *mockModuleRepo) GetAll(ctx context.Context) ([]*content.Module, error) {
	var result []*content.Module
	for _, mod := range m.data {
		result = append(result, mod)
	}
	return result, nil
}

func (m *mockModuleRepo) GetByID(ctx context.Context, id int) (*content.Module, error) {
	mod, ok := m.data[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return mod, nil
}

func (m *mockModuleRepo) Create(ctx context.Context, mod *content.Module) error { return nil }
func (m *mockModuleRepo) Update(ctx context.Context, mod *content.Module) error { return nil }
func (m *mockModuleRepo) Delete(ctx context.Context, id int) error               { return nil }
func (m *mockModuleRepo) GetMaxOrderNum(ctx context.Context) (int, error)        { return 0, nil }

// ── tests ─────────────────────────────────────────────────────────────────────

func TestGetUserProgress_Empty(t *testing.T) {
	uc := NewUseCase(
		newMockProgressRepo(),
		&mockAttemptRepo{},
		&mockTestRepo{},
		newMockThemeRepo(),
		newMockModuleRepo(),
	)

	result, err := uc.GetUserProgress(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalThemes != 0 {
		t.Errorf("TotalThemes = %d, want 0", result.TotalThemes)
	}
	if result.CompletedThemes != 0 {
		t.Errorf("CompletedThemes = %d, want 0", result.CompletedThemes)
	}
	if result.AverageScore != nil {
		t.Error("AverageScore should be nil when no progress")
	}
}

func TestGetUserProgress_WithCompletedThemes(t *testing.T) {
	module := &content.Module{ID: 1, Name: "Anatomy"}
	theme1 := &content.Theme{ID: 10, ModuleID: 1, Name: "Bones"}
	theme2 := &content.Theme{ID: 11, ModuleID: 1, Name: "Muscles"}

	progressRepo := newMockProgressRepo()
	score80 := 80
	score60 := 60
	progressRepo.data[progKey(1, 10)] = &progress.UserProgress{
		UserID: 1, ThemeID: 10, Status: progress.StatusCompleted, Score: &score80,
	}
	progressRepo.data[progKey(1, 11)] = &progress.UserProgress{
		UserID: 1, ThemeID: 11, Status: progress.StatusFailed, Score: &score60,
	}

	uc := NewUseCase(
		progressRepo,
		&mockAttemptRepo{},
		&mockTestRepo{},
		newMockThemeRepo(theme1, theme2),
		newMockModuleRepo(module),
	)

	result, err := uc.GetUserProgress(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalThemes != 2 {
		t.Errorf("TotalThemes = %d, want 2", result.TotalThemes)
	}
	if result.CompletedThemes != 1 {
		t.Errorf("CompletedThemes = %d, want 1", result.CompletedThemes)
	}
	if result.AverageScore == nil || *result.AverageScore != 80 {
		t.Errorf("AverageScore = %v, want 80", result.AverageScore)
	}
}

func TestGetModuleProgress_NotFound(t *testing.T) {
	uc := NewUseCase(
		newMockProgressRepo(),
		&mockAttemptRepo{},
		&mockTestRepo{},
		newMockThemeRepo(),
		newMockModuleRepo(),
	)

	_, err := uc.GetModuleProgress(context.Background(), 1, 999)
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestGetModuleProgress_WithThemes(t *testing.T) {
	module := &content.Module{ID: 2, Name: "Physiology"}
	theme := &content.Theme{ID: 20, ModuleID: 2, Name: "Heart"}

	progressRepo := newMockProgressRepo()
	score := 90
	progressRepo.data[progKey(5, 20)] = &progress.UserProgress{
		UserID: 5, ThemeID: 20, Status: progress.StatusCompleted, Score: &score,
	}

	uc := NewUseCase(
		progressRepo,
		&mockAttemptRepo{},
		&mockTestRepo{},
		newMockThemeRepo(theme),
		newMockModuleRepo(module),
	)

	result, err := uc.GetModuleProgress(context.Background(), 5, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ModuleID != 2 {
		t.Errorf("ModuleID = %d, want 2", result.ModuleID)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
	if result.AverageScore == nil || *result.AverageScore != 90 {
		t.Errorf("AverageScore = %v, want 90", result.AverageScore)
	}
}
