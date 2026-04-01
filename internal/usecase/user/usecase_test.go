package user

import (
	"context"
	"testing"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── mocks ─────────────────────────────────────────────────────────────────────

type mockUserRepo struct {
	data   map[int64]*user.User
	exists map[int64]bool
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		data:   make(map[int64]*user.User),
		exists: make(map[int64]bool),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	m.data[u.TelegramID] = u
	m.exists[u.TelegramID] = true
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
	return m.exists[id], nil
}
func (m *mockUserRepo) GetAll(ctx context.Context, role, subStatus string, limit, offset int) ([]*user.User, int, error) {
	return nil, 0, nil
}

type mockSubRepo struct {
	data map[int64]*subscription.Subscription
}

func newMockSubRepo() *mockSubRepo {
	return &mockSubRepo{data: make(map[int64]*subscription.Subscription)}
}

func (m *mockSubRepo) Create(ctx context.Context, s *subscription.Subscription) error {
	m.data[s.UserID] = s
	return nil
}

func (m *mockSubRepo) GetActiveByUserID(ctx context.Context, userID int64) (*subscription.Subscription, error) {
	s, ok := m.data[userID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return s, nil
}

func (m *mockSubRepo) GetByPaymentID(ctx context.Context, paymentID string) (*subscription.Subscription, error) {
	return nil, apperrors.ErrNotFound
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newUC() *UseCase {
	return NewUseCase(newMockUserRepo(), newMockSubRepo())
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestRegister_HappyPath(t *testing.T) {
	uc := newUC()
	u, err := uc.Register(context.Background(), 100, "ivan_p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.TelegramID != 100 {
		t.Errorf("TelegramID = %d, want 100", u.TelegramID)
	}
	if u.Role != user.RoleStudent {
		t.Errorf("Role = %q, want student", u.Role)
	}
	if u.SubscriptionStatus != user.SubscriptionStatusInactive {
		t.Errorf("SubscriptionStatus = %q, want inactive", u.SubscriptionStatus)
	}
}

func TestRegister_EmptyOptionalFields(t *testing.T) {
	uc := newUC()
	u, err := uc.Register(context.Background(), 101, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Username != nil {
		t.Error("Username should be nil when empty")
	}
}

func TestRegister_AlreadyExists(t *testing.T) {
	uc := newUC()
	_, _ = uc.Register(context.Background(), 200, "")
	_, err := uc.Register(context.Background(), 200, "")
	if err == nil {
		t.Fatal("expected ErrAlreadyExists, got nil")
	}
	if !apperrors.IsConflict(err) {
		t.Errorf("expected conflict error, got %v", err)
	}
}

func TestUpdateRole_HappyPath(t *testing.T) {
	uc := newUC()
	_, _ = uc.Register(context.Background(), 300, "")

	u, err := uc.UpdateRole(context.Background(), 300, user.RoleTeacher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Role != user.RoleTeacher {
		t.Errorf("Role = %q, want teacher", u.Role)
	}
}

func TestUpdateRole_NotFound(t *testing.T) {
	uc := newUC()
	_, err := uc.UpdateRole(context.Background(), 9999, user.RoleTeacher)
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestUpdateSettings_Language(t *testing.T) {
	uc := newUC()
	_, _ = uc.Register(context.Background(), 400, "")

	lang := "en"
	u, err := uc.UpdateSettings(context.Background(), 400, &lang, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Language != "en" {
		t.Errorf("Language = %q, want en", u.Language)
	}
}

func TestUpdateSettings_Notifications(t *testing.T) {
	uc := newUC()
	_, _ = uc.Register(context.Background(), 401, "")

	disabled := false
	u, err := uc.UpdateSettings(context.Background(), 401, nil, &disabled)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.NotificationsEnabled {
		t.Error("NotificationsEnabled should be false")
	}
}

func TestGetSubscription_NotFound(t *testing.T) {
	uc := newUC()
	_, _ = uc.Register(context.Background(), 500, "")

	_, err := uc.GetSubscription(context.Background(), 500)
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestGetSubscription_UserNotFound(t *testing.T) {
	uc := newUC()
	_, err := uc.GetSubscription(context.Background(), 9999)
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not found, got %v", err)
	}
}
