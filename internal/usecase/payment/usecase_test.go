package payment

import (
	"context"
	"testing"

	"github.com/vladkonst/mnemonics/internal/domain/subscription"
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

type mockSubRepo struct {
	data      map[int64]*subscription.Subscription
	byPayment map[string]*subscription.Subscription
}

func newMockSubRepo() *mockSubRepo {
	return &mockSubRepo{
		data:      make(map[int64]*subscription.Subscription),
		byPayment: make(map[string]*subscription.Subscription),
	}
}

func (m *mockSubRepo) Create(ctx context.Context, s *subscription.Subscription) error {
	m.data[s.UserID] = s
	m.byPayment[s.PaymentID] = s
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
	s, ok := m.byPayment[paymentID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return s, nil
}

type mockPaymentService struct {
	invoiceID  string
	paymentURL string
	amount     int
}

func (m *mockPaymentService) CreateInvoice(ctx context.Context, userID int64, plan string) (string, string, int, error) {
	return m.invoiceID, m.paymentURL, m.amount, nil
}

func (m *mockPaymentService) VerifyWebhookSignature(payload []byte, signature string) error {
	return nil
}

type mockNotificationService struct{}

func (m *mockNotificationService) Send(ctx context.Context, id int64, msg string) error {
	return nil
}

func newUC() (*UseCase, *mockUserRepo, *mockSubRepo) {
	userRepo := newMockUserRepo()
	subRepo := newMockSubRepo()
	paymentSvc := &mockPaymentService{invoiceID: "inv_123", paymentURL: "https://pay.example.com", amount: 990}
	return NewUseCase(userRepo, subRepo, paymentSvc, &mockNotificationService{}), userRepo, subRepo
}

func addUser(repo *mockUserRepo, id int64) {
	repo.data[id] = &user.User{
		TelegramID:         id,
		Role:               user.RoleStudent,
		SubscriptionStatus: user.SubscriptionStatusInactive,
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestCreateInvoice_HappyPath(t *testing.T) {
	uc, userRepo, _ := newUC()
	addUser(userRepo, 100)

	result, err := uc.CreateInvoice(context.Background(), 100, "monthly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.InvoiceID != "inv_123" {
		t.Errorf("InvoiceID = %q, want inv_123", result.InvoiceID)
	}
	if result.Plan != "monthly" {
		t.Errorf("Plan = %q, want monthly", result.Plan)
	}

	// User should have pending payment set
	u, _ := userRepo.GetByID(context.Background(), 100)
	if u.PendingPaymentID == nil || *u.PendingPaymentID != "inv_123" {
		t.Error("PendingPaymentID should be set on user")
	}
}

func TestCreateInvoice_UserNotFound(t *testing.T) {
	uc, _, _ := newUC()

	_, err := uc.CreateInvoice(context.Background(), 9999, "monthly")
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestCreateInvoice_AlreadyActiveSubscription(t *testing.T) {
	uc, userRepo, subRepo := newUC()
	addUser(userRepo, 200)
	subRepo.data[200] = &subscription.Subscription{
		UserID: 200,
		Status: subscription.SubscriptionPlanStatusActive,
	}

	_, err := uc.CreateInvoice(context.Background(), 200, "monthly")
	if !apperrors.IsConflict(err) {
		t.Errorf("expected conflict, got %v", err)
	}
}

func TestHandleWebhook_Succeeded(t *testing.T) {
	uc, userRepo, subRepo := newUC()
	addUser(userRepo, 300)

	event := WebhookEvent{
		PaymentID: "pay_abc",
		UserID:    300,
		Plan:      "monthly",
		Status:    "succeeded",
	}
	err := uc.HandleWebhook(context.Background(), []byte(`{}`), "", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Subscription should be created
	sub, err := subRepo.GetByPaymentID(context.Background(), "pay_abc")
	if err != nil {
		t.Fatalf("subscription not created: %v", err)
	}
	if sub.Status != subscription.SubscriptionPlanStatusActive {
		t.Errorf("sub.Status = %q, want active", sub.Status)
	}

	// User should have active subscription
	u, _ := userRepo.GetByID(context.Background(), 300)
	if u.SubscriptionStatus != user.SubscriptionStatusActive {
		t.Errorf("user.SubscriptionStatus = %q, want active", u.SubscriptionStatus)
	}
}

func TestHandleWebhook_NonSucceeded(t *testing.T) {
	uc, userRepo, subRepo := newUC()
	addUser(userRepo, 400)

	event := WebhookEvent{PaymentID: "pay_fail", UserID: 400, Plan: "monthly", Status: "cancelled"}
	err := uc.HandleWebhook(context.Background(), []byte(`{}`), "", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No subscription should be created
	if _, ok := subRepo.byPayment["pay_fail"]; ok {
		t.Error("subscription should not be created for cancelled payment")
	}
}

func TestHandleWebhook_Idempotent(t *testing.T) {
	uc, userRepo, _ := newUC()
	addUser(userRepo, 500)

	event := WebhookEvent{PaymentID: "pay_dup", UserID: 500, Plan: "monthly", Status: "succeeded"}

	// Process twice
	_ = uc.HandleWebhook(context.Background(), []byte(`{}`), "", event)
	err := uc.HandleWebhook(context.Background(), []byte(`{}`), "", event)
	if err != nil {
		t.Fatalf("second webhook call should be idempotent, got: %v", err)
	}
}
