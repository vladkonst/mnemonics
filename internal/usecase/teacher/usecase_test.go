package teacher

import (
	"context"
	"fmt"
	"testing"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── mocks ─────────────────────────────────────────────────────────────────────

type mockUserRepo struct {
	data map[int64]*user.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{data: make(map[int64]*user.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	m.data[u.TelegramID] = u
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	u, ok := m.data[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error {
	m.data[u.TelegramID] = u
	return nil
}

func (m *mockUserRepo) Exists(ctx context.Context, id int64) (bool, error) {
	_, ok := m.data[id]
	return ok, nil
}
func (m *mockUserRepo) GetAll(ctx context.Context, role, subStatus string, limit, offset int) ([]*user.User, int, error) {
	return nil, 0, nil
}

type mockTeacherStudentRepo struct {
	students map[int64][]*user.User
	pairs    map[string]bool
}

func newMockTeacherStudentRepo() *mockTeacherStudentRepo {
	return &mockTeacherStudentRepo{
		students: make(map[int64][]*user.User),
		pairs:    make(map[string]bool),
	}
}

func pairKey(tID, sID int64) string {
	return fmt.Sprintf("%d-%d", tID, sID)
}

func (m *mockTeacherStudentRepo) AddStudent(ctx context.Context, teacherID, studentID int64, code string) error {
	// lookup user from somewhere — for test just skip
	return nil
}

func (m *mockTeacherStudentRepo) GetStudentsByTeacher(ctx context.Context, teacherID int64) ([]*user.User, error) {
	return m.students[teacherID], nil
}

func (m *mockTeacherStudentRepo) IsTeacherStudent(ctx context.Context, teacherID, studentID int64) (bool, error) {
	return m.pairs[pairKey(teacherID, studentID)], nil
}

type mockProgressRepo struct {
	data map[string]*progress.UserProgress
}

func newMockProgressRepo() *mockProgressRepo {
	return &mockProgressRepo{data: make(map[string]*progress.UserProgress)}
}

func (m *mockProgressRepo) Upsert(ctx context.Context, p *progress.UserProgress) error { return nil }

func (m *mockProgressRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) (*progress.UserProgress, error) {
	return nil, apperrors.ErrNotFound
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
	return 0, nil
}

type mockAttemptRepo struct{}

func (m *mockAttemptRepo) Create(ctx context.Context, a *progress.TestAttempt) error { return nil }
func (m *mockAttemptRepo) GetByAttemptID(ctx context.Context, id string) (*progress.TestAttempt, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockAttemptRepo) GetByUserAndTheme(ctx context.Context, userID int64, themeID int) ([]*progress.TestAttempt, error) {
	return nil, nil
}

type mockModuleRepo struct {
	data map[int]*content.Module
}

func newMockModuleRepo() *mockModuleRepo {
	return &mockModuleRepo{data: make(map[int]*content.Module)}
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

type mockThemeRepo struct {
	data map[int]*content.Theme
}

func newMockThemeRepo() *mockThemeRepo {
	return &mockThemeRepo{data: make(map[int]*content.Theme)}
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

// ── helpers ───────────────────────────────────────────────────────────────────

func newUC(userRepo *mockUserRepo, tsRepo *mockTeacherStudentRepo) *UseCase {
	return NewUseCase(tsRepo, newMockProgressRepo(), &mockAttemptRepo{}, newMockModuleRepo(), newMockThemeRepo(), userRepo)
}

func makeTeacher(id int64) *user.User {
	return &user.User{TelegramID: id, Role: user.RoleTeacher}
}

func makeStudent(id int64) *user.User {
	return &user.User{TelegramID: id, Role: user.RoleStudent}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestGetStudents_HappyPath(t *testing.T) {
	userRepo := newMockUserRepo()
	tsRepo := newMockTeacherStudentRepo()

	teacher := makeTeacher(1)
	student := makeStudent(2)
	userRepo.data[1] = teacher
	userRepo.data[2] = student
	tsRepo.students[1] = []*user.User{student}

	uc := newUC(userRepo, tsRepo)
	result, err := uc.GetStudents(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}

func TestGetStudents_NotTeacher(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.data[1] = makeStudent(1) // not a teacher
	tsRepo := newMockTeacherStudentRepo()

	uc := newUC(userRepo, tsRepo)
	_, err := uc.GetStudents(context.Background(), 1)
	if !apperrors.IsForbidden(err) {
		t.Errorf("expected forbidden, got %v", err)
	}
}

func TestGetStudents_NotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	tsRepo := newMockTeacherStudentRepo()

	uc := newUC(userRepo, tsRepo)
	_, err := uc.GetStudents(context.Background(), 9999)
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestGetStudentProgress_NotYourStudent(t *testing.T) {
	userRepo := newMockUserRepo()
	tsRepo := newMockTeacherStudentRepo()

	userRepo.data[1] = makeTeacher(1)
	userRepo.data[2] = makeStudent(2)
	// student NOT in teacher's list

	uc := newUC(userRepo, tsRepo)
	_, err := uc.GetStudentProgress(context.Background(), 1, 2)
	if !apperrors.IsForbidden(err) {
		t.Errorf("expected forbidden (not your student), got %v", err)
	}
}

func TestGetStudentProgress_HappyPath(t *testing.T) {
	userRepo := newMockUserRepo()
	tsRepo := newMockTeacherStudentRepo()

	userRepo.data[1] = makeTeacher(1)
	userRepo.data[2] = makeStudent(2)
	tsRepo.pairs[pairKey(1, 2)] = true

	uc := newUC(userRepo, tsRepo)
	result, err := uc.GetStudentProgress(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Student.TelegramID != 2 {
		t.Errorf("student ID = %d, want 2", result.Student.TelegramID)
	}
}

func TestGetStatistics_NoStudents(t *testing.T) {
	userRepo := newMockUserRepo()
	tsRepo := newMockTeacherStudentRepo()
	userRepo.data[1] = makeTeacher(1)

	uc := newUC(userRepo, tsRepo)
	result, err := uc.GetStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalStudents != 0 {
		t.Errorf("TotalStudents = %d, want 0", result.TotalStudents)
	}
}

func TestGetStatistics_NotTeacher(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.data[1] = makeStudent(1)
	tsRepo := newMockTeacherStudentRepo()

	uc := newUC(userRepo, tsRepo)
	_, err := uc.GetStatistics(context.Background(), 1)
	if !apperrors.IsForbidden(err) {
		t.Errorf("expected forbidden, got %v", err)
	}
}
