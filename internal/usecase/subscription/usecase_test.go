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

type mockPromoCodeRepo struct {
	codes map[string]*subscription.PromoCode
}

func (m *mockPromoCodeRepo) GetByCode(ctx context.Context, code string) (*subscription.PromoCode, error) {
	p, ok := m.codes[code]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return p, nil
}
func (m *mockPromoCodeRepo) Update(ctx context.Context, p *subscription.PromoCode) error {
	m.codes[p.Code] = p
	return nil
}
func (m *mockPromoCodeRepo) Create(ctx context.Context, p *subscription.PromoCode) error {
	m.codes[p.Code] = p
	return nil
}
func (m *mockPromoCodeRepo) Deactivate(ctx context.Context, code string) error {
	p, ok := m.codes[code]
	if !ok {
		return apperrors.ErrNotFound
	}
	p.Deactivate()
	return nil
}
func (m *mockPromoCodeRepo) GetByTeacherID(ctx context.Context, teacherID int64) ([]*subscription.PromoCode, error) {
	var result []*subscription.PromoCode
	for _, p := range m.codes {
		if p.TeacherID != nil && *p.TeacherID == teacherID {
			result = append(result, p)
		}
	}
	return result, nil
}
func (m *mockPromoCodeRepo) ConsumeOne(ctx context.Context, code string) error {
	p, ok := m.codes[code]
	if !ok {
		return apperrors.ErrNotFound
	}
	if p.Remaining <= 0 {
		return apperrors.ErrPromoCodeExhausted
	}
	p.Remaining--
	return nil
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

type mockTeacherStudentRepo struct {
	relationships map[string]bool // "teacherID-studentID"
}

func (m *mockTeacherStudentRepo) AddStudent(ctx context.Context, teacherID, studentID int64, promoCode string) error {
	key := teacherStudentKey(teacherID, studentID)
	m.relationships[key] = true
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

type mockNotificationService struct {
	sent []string
}

func (m *mockNotificationService) Send(ctx context.Context, telegramID int64, message string) error {
	m.sent = append(m.sent, message)
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newUseCase(
	promoCodes *mockPromoCodeRepo,
	subs *mockSubscriptionRepo,
	users *mockUserRepo,
	teacherStudents *mockTeacherStudentRepo,
) *ucSub.UseCase {
	return ucSub.NewUseCase(promoCodes, subs, users, teacherStudents, &mockNotificationService{})
}

// ── Tests: ActivatePromoCode ──────────────────────────────────────────────────

func TestActivatePromoCode_HappyPath(t *testing.T) {
	teacherID := int64(3001)
	code := "UNIV-CODE-001"

	users := &mockUserRepo{users: map[int64]*user.User{
		teacherID: {TelegramID: teacherID, Role: user.RoleTeacher},
	}}
	promos := &mockPromoCodeRepo{codes: map[string]*subscription.PromoCode{
		code: {
			Code:           code,
			UniversityName: "Test University",
			MaxActivations: 50,
			Remaining:      50,
			Status:         subscription.PromoCodeStatusPending,
		},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}

	uc := newUseCase(promos, subs, users, ts)

	promo, err := uc.ActivatePromoCode(context.Background(), teacherID, code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promo.Status != subscription.PromoCodeStatusActive {
		t.Errorf("expected status=active, got %s", promo.Status)
	}
	if promo.TeacherID == nil || *promo.TeacherID != teacherID {
		t.Errorf("expected TeacherID=%d", teacherID)
	}
}

func TestActivatePromoCode_PromoNotFound(t *testing.T) {
	teacherID := int64(3002)

	users := &mockUserRepo{users: map[int64]*user.User{
		teacherID: {TelegramID: teacherID, Role: user.RoleTeacher},
	}}
	promos := &mockPromoCodeRepo{codes: map[string]*subscription.PromoCode{}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}

	uc := newUseCase(promos, subs, users, ts)

	_, err := uc.ActivatePromoCode(context.Background(), teacherID, "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent promo code, got nil")
	}
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestActivatePromoCode_AlreadyActivated(t *testing.T) {
	teacherID := int64(3003)
	otherTeacherID := int64(9999)
	code := "ALREADY-ACTIVE"

	users := &mockUserRepo{users: map[int64]*user.User{
		teacherID: {TelegramID: teacherID, Role: user.RoleTeacher},
	}}
	promos := &mockPromoCodeRepo{codes: map[string]*subscription.PromoCode{
		code: {
			Code:           code,
			UniversityName: "University",
			MaxActivations: 30,
			Remaining:      30,
			Status:         subscription.PromoCodeStatusActive, // already activated
			TeacherID:      &otherTeacherID,
		},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}

	uc := newUseCase(promos, subs, users, ts)

	_, err := uc.ActivatePromoCode(context.Background(), teacherID, code)
	if err == nil {
		t.Fatal("expected error for already-activated promo code, got nil")
	}
	if err != apperrors.ErrAlreadyActivated {
		t.Errorf("expected ErrAlreadyActivated, got %v", err)
	}
}

func TestActivatePromoCode_NotTeacher(t *testing.T) {
	userID := int64(3004)
	code := "PROMO-001"

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID, Role: user.RoleStudent},
	}}
	promos := &mockPromoCodeRepo{codes: map[string]*subscription.PromoCode{
		code: {
			Code:   code,
			Status: subscription.PromoCodeStatusPending,
		},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}

	uc := newUseCase(promos, subs, users, ts)

	_, err := uc.ActivatePromoCode(context.Background(), userID, code)
	if err == nil {
		t.Fatal("expected ErrNotTeacher, got nil")
	}
	if err != apperrors.ErrNotTeacher {
		t.Errorf("expected ErrNotTeacher, got %v", err)
	}
}

// ── Tests: CreatePromoSubscription ───────────────────────────────────────────

func TestCreatePromoSubscription_HappyPath(t *testing.T) {
	userID := int64(4001)
	teacherID := int64(4000)
	code := "PROMO-HAPPY"

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID, Role: user.RoleStudent, SubscriptionStatus: user.SubscriptionStatusInactive},
	}}

	futureExpiry := time.Now().UTC().Add(30 * 24 * time.Hour)
	promos := &mockPromoCodeRepo{codes: map[string]*subscription.PromoCode{
		code: {
			Code:           code,
			UniversityName: "University",
			MaxActivations: 10,
			Remaining:      10,
			Status:         subscription.PromoCodeStatusActive,
			TeacherID:      &teacherID,
			ExpiresAt:      &futureExpiry,
		},
	}}
	subs := &mockSubscriptionRepo{
		active:    map[int64]*subscription.Subscription{},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}

	uc := newUseCase(promos, subs, users, ts)

	sub, err := uc.CreatePromoSubscription(context.Background(), userID, code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.UserID != userID {
		t.Errorf("expected UserID=%d, got %d", userID, sub.UserID)
	}
	if sub.Type != subscription.SubscriptionTypeUniversity {
		t.Errorf("expected type=university, got %s", sub.Type)
	}

	// Verify promo remaining was decremented.
	promo := promos.codes[code]
	if promo.Remaining != 9 {
		t.Errorf("expected Remaining=9, got %d", promo.Remaining)
	}
}

func TestCreatePromoSubscription_AlreadyHasSubscription(t *testing.T) {
	userID := int64(4002)

	users := &mockUserRepo{users: map[int64]*user.User{
		userID: {TelegramID: userID, Role: user.RoleStudent, SubscriptionStatus: user.SubscriptionStatusActive},
	}}
	promos := &mockPromoCodeRepo{codes: map[string]*subscription.PromoCode{}}
	subs := &mockSubscriptionRepo{
		active: map[int64]*subscription.Subscription{
			userID: {
				PaymentID: "existing",
				UserID:    userID,
				Status:    subscription.SubscriptionPlanStatusActive,
			},
		},
		byPayment: map[string]*subscription.Subscription{},
	}
	ts := &mockTeacherStudentRepo{relationships: map[string]bool{}}

	uc := newUseCase(promos, subs, users, ts)

	_, err := uc.CreatePromoSubscription(context.Background(), userID, "ANY-CODE")
	if err == nil {
		t.Fatal("expected ErrActiveSubscriptionExists, got nil")
	}
	if err != apperrors.ErrActiveSubscriptionExists {
		t.Errorf("expected ErrActiveSubscriptionExists, got %v", err)
	}
}
