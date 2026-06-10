package subscription_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	ucSub "github.com/vladkonst/mnemonics/internal/usecase/subscription"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── Hand-written mocks ────────────────────────────────────────────────────────

type mockUserRepo struct {
	users map[int64]*user.User
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	m.users[u.TelegramID] = u
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return u, nil
}
func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error {
	m.users[u.TelegramID] = u
	return nil
}
func (m *mockUserRepo) Exists(ctx context.Context, id int64) (bool, error) {
	_, ok := m.users[id]
	return ok, nil
}
func (m *mockUserRepo) GetAll(ctx context.Context, role, subStatus string, limit, offset int) ([]*user.User, int, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id int64) error { return nil }

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

type mockTeacherStudentRepo struct {
	relationships map[string]bool
}

func (m *mockTeacherStudentRepo) AddStudent(ctx context.Context, teacherID, studentID int64, joinRef string) error {
	m.relationships[teacherStudentKey(teacherID, studentID)] = true
	return nil
}
func (m *mockTeacherStudentRepo) GetStudentsByTeacher(ctx context.Context, teacherID int64) ([]*user.User, error) {
	return nil, nil
}
func (m *mockTeacherStudentRepo) IsTeacherStudent(ctx context.Context, teacherID, studentID int64) (bool, error) {
	return m.relationships[teacherStudentKey(teacherID, studentID)], nil
}

func teacherStudentKey(teacherID, studentID int64) string {
	return fmt.Sprintf("%d-%d", teacherID, studentID)
}

type mockInviteLinkRepo struct {
	links       map[string]*subscription.InviteLink
	activations map[string]int // linkID → count
}

func (m *mockInviteLinkRepo) Create(ctx context.Context, l *subscription.InviteLink) error {
	m.links[l.ID] = l
	return nil
}
func (m *mockInviteLinkRepo) GetByID(ctx context.Context, id string) (*subscription.InviteLink, error) {
	l, ok := m.links[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return l, nil
}
func (m *mockInviteLinkRepo) GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.InviteLink, error) {
	return nil, nil
}
func (m *mockInviteLinkRepo) GetAll(ctx context.Context, limit, offset int) ([]*subscription.InviteLink, int, error) {
	return nil, 0, nil
}
func (m *mockInviteLinkRepo) AddActivation(ctx context.Context, a *subscription.InviteLinkActivation) error {
	m.activations[a.InviteLinkID]++
	return nil
}
func (m *mockInviteLinkRepo) GetActivations(ctx context.Context, linkID string) ([]*subscription.InviteLinkActivation, error) {
	return nil, nil
}
func (m *mockInviteLinkRepo) CountActivations(ctx context.Context, linkID string) (int, error) {
	return m.activations[linkID], nil
}

type mockNotificationService struct{}

func (m *mockNotificationService) Send(_ context.Context, _ int64, _ string) error    { return nil }
func (m *mockNotificationService) SendDocument(_ context.Context, _ int64, _ string, _ []byte, _ string) error {
	return nil
}

// ── Helper ────────────────────────────────────────────────────────────────────

func newUseCase(
	subs *mockSubscriptionRepo,
	users *mockUserRepo,
	ts *mockTeacherStudentRepo,
	links *mockInviteLinkRepo,
) *ucSub.UseCase {
	return ucSub.NewUseCase(subs, users, ts, links, &mockNotificationService{}, nil)
}

// ── Tests: CreateInviteSubscription ──────────────────────────────────────────

func TestCreateInviteSubscription_HappyPath(t *testing.T) {
	userID := int64(1001)
	teacherID := int64(1000)
	linkID := "link-abc"

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID, Role: user.RoleStudent, SubscriptionStatus: user.SubscriptionStatusInactive},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}
	links := &mockInviteLinkRepo{
		links: map[string]*subscription.InviteLink{
			linkID: {ID: linkID, TeacherID: teacherID, MaxActivations: 30, CreatedAt: time.Now()},
		},
		activations: map[string]int{},
	}

	uc := newUseCase(subs, users, ts, links)

	sub, err := uc.CreateInviteSubscription(context.Background(), userID, linkID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.UserID != userID {
		t.Errorf("expected UserID=%d, got %d", userID, sub.UserID)
	}
	if sub.Type != subscription.SubscriptionTypeUniversity {
		t.Errorf("expected type=university, got %s", sub.Type)
	}
	if links.activations[linkID] != 1 {
		t.Errorf("expected 1 activation recorded, got %d", links.activations[linkID])
	}
	if !ts.relationships[teacherStudentKey(teacherID, userID)] {
		t.Error("expected teacher-student relationship to be recorded")
	}
}

func TestCreateInviteSubscription_LinkExhausted(t *testing.T) {
	userID := int64(1002)
	linkID := "link-full"

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID, Role: user.RoleStudent},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}
	links := &mockInviteLinkRepo{
		links: map[string]*subscription.InviteLink{
			linkID: {ID: linkID, TeacherID: 999, MaxActivations: 2, CreatedAt: time.Now()},
		},
		activations: map[string]int{linkID: 2}, // already full
	}

	uc := newUseCase(subs, users, ts, links)

	_, err := uc.CreateInviteSubscription(context.Background(), userID, linkID)
	if err == nil {
		t.Fatal("expected ErrInviteLinkExhausted, got nil")
	}
	if err != apperrors.ErrInviteLinkExhausted {
		t.Errorf("expected ErrInviteLinkExhausted, got %v", err)
	}
}

func TestCreateInviteSubscription_LinkNotFound(t *testing.T) {
	userID := int64(1003)

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	links := &mockInviteLinkRepo{links: map[string]*subscription.InviteLink{}, activations: map[string]int{}}

	uc := newUseCase(subs, users, &mockTeacherStudentRepo{relationships: map[string]bool{}}, links)

	_, err := uc.CreateInviteSubscription(context.Background(), userID, "nonexistent")
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateInviteSubscription_AlreadyHasSubscription(t *testing.T) {
	userID := int64(1004)
	linkID := "link-ok"

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID, SubscriptionStatus: user.SubscriptionStatusActive},
	}}
	subs := &mockSubscriptionRepo{
		active: map[int64]*subscription.Subscription{
			userID: {PaymentID: "existing", UserID: userID, Status: subscription.SubscriptionPlanStatusActive},
		},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}
	links := &mockInviteLinkRepo{
		links:       map[string]*subscription.InviteLink{linkID: {ID: linkID, TeacherID: 999, MaxActivations: 30}},
		activations: map[string]int{},
	}

	uc := newUseCase(subs, users, ts, links)

	_, err := uc.CreateInviteSubscription(context.Background(), userID, linkID)
	if err != apperrors.ErrActiveSubscriptionExists {
		t.Errorf("expected ErrActiveSubscriptionExists, got %v", err)
	}
}

// ── Tests: CreatePaymentSubscription ─────────────────────────────────────────

func TestCreatePaymentSubscription_HappyPath(t *testing.T) {
	userID := int64(2001)
	paymentID := "pay-001"

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID, Role: user.RoleStudent, SubscriptionStatus: user.SubscriptionStatusInactive},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	links := &mockInviteLinkRepo{links: map[string]*subscription.InviteLink{}, activations: map[string]int{}}

	uc := newUseCase(subs, users, &mockTeacherStudentRepo{relationships: map[string]bool{}}, links)

	sub, err := uc.CreatePaymentSubscription(context.Background(), userID, paymentID, "monthly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.PaymentID != paymentID {
		t.Errorf("expected paymentID=%s, got %s", paymentID, sub.PaymentID)
	}
	if sub.Type != subscription.SubscriptionTypePersonal {
		t.Errorf("expected type=personal, got %s", sub.Type)
	}
}

func TestCreatePaymentSubscription_Idempotent(t *testing.T) {
	userID := int64(2002)
	paymentID := "pay-idem"

	existing := &subscription.Subscription{PaymentID: paymentID, UserID: userID, Status: subscription.SubscriptionPlanStatusActive}
	users := &mockUserRepo{users: map[int64]*user.User{userID: {TelegramID: userID}}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{userID: existing},
		byPayment: map[string]*subscription.Subscription{paymentID: existing},
	}
	links := &mockInviteLinkRepo{links: map[string]*subscription.InviteLink{}, activations: map[string]int{}}

	uc := newUseCase(subs, users, &mockTeacherStudentRepo{relationships: map[string]bool{}}, links)

	sub, err := uc.CreatePaymentSubscription(context.Background(), userID, paymentID, "monthly")
	if err != nil {
		t.Fatalf("unexpected error on idempotent call: %v", err)
	}
	if sub.PaymentID != paymentID {
		t.Errorf("expected existing sub returned")
	}
}
