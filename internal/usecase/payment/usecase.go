// Package payment provides use cases for payment invoice creation and webhook handling.
package payment

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	pdfpkg "github.com/vladkonst/mnemonics/internal/infrastructure/pdf"
	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// InvoiceResult is returned after creating a payment invoice.
type InvoiceResult struct {
	InvoiceID  string
	PaymentURL string
	Amount     int
	Plan       string
}

// WebhookEvent represents a parsed event from the payment gateway.
type WebhookEvent struct {
	PaymentID string
	UserID    int64
	Plan      string
	Status    string // "succeeded", "cancelled", etc.
}

// UseCase orchestrates payment operations.
type UseCase struct {
	users           interfaces.UserRepository
	subscriptions   interfaces.SubscriptionRepository
	payment         interfaces.PaymentService
	notifications   interfaces.NotificationService
	corporateGroups interfaces.CorporateGroupRepository
	inviteLinks     interfaces.InviteLinkRepository
	botUsername     string
}

// NewUseCase creates a new payment UseCase.
func NewUseCase(
	users interfaces.UserRepository,
	subscriptions interfaces.SubscriptionRepository,
	payment interfaces.PaymentService,
	notifications interfaces.NotificationService,
	corporateGroups interfaces.CorporateGroupRepository,
	inviteLinks interfaces.InviteLinkRepository,
	botUsername string,
) *UseCase {
	return &UseCase{
		users:           users,
		subscriptions:   subscriptions,
		payment:         payment,
		notifications:   notifications,
		corporateGroups: corporateGroups,
		inviteLinks:     inviteLinks,
		botUsername:     botUsername,
	}
}

// CreateInvoice calls the payment gateway and records a pending payment on the user.
func (uc *UseCase) CreateInvoice(ctx context.Context, userID int64, plan string) (*InvoiceResult, error) {
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Managers can hold multiple corporate subscriptions — skip the uniqueness check.
	if u.Role != user.RoleManager {
		active, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
		if err != nil && !apperrors.IsNotFound(err) {
			return nil, err
		}
		if active != nil && active.IsActive() {
			return nil, apperrors.ErrActiveSubscriptionExists
		}
	}

	invoiceID, paymentURL, amount, err := uc.payment.CreateInvoice(ctx, userID, plan)
	if err != nil {
		return nil, fmt.Errorf("payment gateway: %w", err)
	}

	// Record the pending payment on the user.
	u.SetPendingPayment(invoiceID)
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	return &InvoiceResult{
		InvoiceID:  invoiceID,
		PaymentURL: paymentURL,
		Amount:     amount,
		Plan:       plan,
	}, nil
}

// HandleWebhook verifies the signature and idempotently processes a payment event.
func (uc *UseCase) HandleWebhook(ctx context.Context, payload []byte, signature string, event WebhookEvent) error {
	// Verify signature first.
	if err := uc.payment.VerifyWebhookSignature(payload, signature); err != nil {
		return fmt.Errorf("%w: %v", apperrors.ErrForbidden, err)
	}

	if event.Status != "succeeded" {
		// Nothing to do for non-success events.
		return nil
	}

	// Idempotency: check if subscription already created for this payment.
	existing, err := uc.subscriptions.GetByPaymentID(ctx, event.PaymentID)
	if err != nil && !apperrors.IsNotFound(err) {
		return err
	}
	if existing != nil {
		// Already processed.
		return nil
	}

	// Determine plan duration.
	plan := event.Plan
	if plan == "" {
		plan = "monthly"
	}

	// Corporate plans are handled separately.
	if strings.HasPrefix(plan, "corporate:") {
		return uc.handleCorporateWebhook(ctx, event, plan)
	}

	// Activate subscription via CreatePaymentSubscription-like logic inline.
	// We do it here directly to avoid circular use case dependencies.
	u, err := uc.users.GetByID(ctx, event.UserID)
	if err != nil {
		return err
	}

	planCopy := plan
	sub := &subscription.Subscription{
		PaymentID: event.PaymentID,
		UserID:    event.UserID,
		Type:      subscription.SubscriptionTypePersonal,
		Status:    subscription.SubscriptionPlanStatusActive,
		Plan:      &planCopy,
		AutoRenew: false,
	}
	if err := uc.subscriptions.Create(ctx, sub); err != nil {
		return err
	}

	u.ActivateSubscription(nil)
	u.ClearPendingPayment()
	if u.Role == user.RoleUnknown {
		u.SetRole(user.RoleStudent)
	}
	if err := uc.users.Update(ctx, u); err != nil {
		return err
	}

	// Notify user.
	_ = uc.notifications.Send(ctx, event.UserID, "Оплата прошла успешно! Подписка активирована.")

	return nil
}

// CreateCorporatePurchaseDirect creates a corporate purchase and all groups/links
// without going through the payment gateway. Returns the new paymentID.
func (uc *UseCase) CreateCorporatePurchaseDirect(ctx context.Context, userID int64, numGroups, numSemesters int) (string, error) {
	if numGroups < 1 || numSemesters < 1 {
		return "", apperrors.ErrInvalidInput
	}

	// Promote user to manager role if not already set.
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if u.Role != user.RoleManager {
		u.SetRole(user.RoleManager)
		if err := uc.users.Update(ctx, u); err != nil {
			return "", err
		}
	}

	now := time.Now().UTC()
	paymentID := uuid.New().String()

	purchase := &subscription.CorporatePurchase{
		PaymentID:   paymentID,
		ManagerID:   userID,
		GroupsCount: numGroups,
		Semesters:   numSemesters,
		TotalAmount: 4470 * numGroups * numSemesters,
		CreatedAt:   now,
	}
	if err := uc.corporateGroups.CreatePurchase(ctx, purchase); err != nil {
		return "", err
	}

	for i := 1; i <= numGroups; i++ {
		linkID := uuid.New().String()
		if err := uc.inviteLinks.Create(ctx, &subscription.InviteLink{
			ID:             linkID,
			TeacherID:      userID,
			MaxActivations: 30,
			CreatedAt:      now,
		}); err != nil {
			return "", fmt.Errorf("create invite link group %d: %w", i, err)
		}

		if err := uc.corporateGroups.Create(ctx, &subscription.CorporateGroup{
			ID:              uuid.New().String(),
			Name:            fmt.Sprintf("Группа %d", i),
			PurchaseID:      paymentID,
			ManagerID:       userID,
			TeacherJoinCode: uuid.New().String(),
			StudentLinkID:   linkID,
			Semesters:       numSemesters,
			CreatedAt:       now,
		}); err != nil {
			return "", fmt.Errorf("create group %d: %w", i, err)
		}
	}

	return paymentID, nil
}

// ── Corporate webhook handling ────────────────────────────────────────────────

// handleCorporateWebhook processes a webhook for a corporate plan purchase.
// Plan format: "corporate:G:S" where G = number of groups, S = semesters.
func (uc *UseCase) handleCorporateWebhook(ctx context.Context, event WebhookEvent, plan string) error {
	parts := strings.SplitN(plan, ":", 3)
	if len(parts) != 3 {
		return fmt.Errorf("invalid corporate plan format: %q", plan)
	}
	numGroups, err := strconv.Atoi(parts[1])
	if err != nil || numGroups < 1 {
		return fmt.Errorf("invalid group count in plan %q", plan)
	}
	numSemesters, err := strconv.Atoi(parts[2])
	if err != nil || numSemesters < 1 {
		return fmt.Errorf("invalid semester count in plan %q", plan)
	}

	// Idempotency: if purchase already exists, skip creation and just resend PDF.
	existingPurchase, err := uc.corporateGroups.GetPurchaseByPaymentID(ctx, event.PaymentID)
	if err != nil && !apperrors.IsNotFound(err) {
		return err
	}
	if existingPurchase != nil {
		return uc.sendCorporatePDF(ctx, event.UserID, event.PaymentID)
	}

	manager, err := uc.users.GetByID(ctx, event.UserID)
	if err != nil {
		return err
	}

	// Create a subscription record for idempotency tracking (same table as personal subs).
	now := time.Now().UTC()
	planCopy := plan
	sub := &subscription.Subscription{
		PaymentID: event.PaymentID,
		UserID:    event.UserID,
		Type:      subscription.SubscriptionTypePersonal,
		Status:    subscription.SubscriptionPlanStatusActive,
		Plan:      &planCopy,
		CreatedAt: now,
	}
	if err := uc.subscriptions.Create(ctx, sub); err != nil {
		return err
	}
	manager.ClearPendingPayment()
	if err := uc.users.Update(ctx, manager); err != nil {
		return err
	}

	// Create the purchase record.
	totalAmount := 4470 * numGroups * numSemesters
	purchase := &subscription.CorporatePurchase{
		PaymentID:   event.PaymentID,
		ManagerID:   event.UserID,
		GroupsCount: numGroups,
		Semesters:   numSemesters,
		TotalAmount: totalAmount,
		CreatedAt:   now,
	}
	if err := uc.corporateGroups.CreatePurchase(ctx, purchase); err != nil {
		return err
	}

	// Create N groups, each with an invite link.
	for i := 1; i <= numGroups; i++ {
		linkID := uuid.New().String()
		link := &subscription.InviteLink{
			ID:             linkID,
			TeacherID:      event.UserID,
			MaxActivations: 30,
			CreatedAt:      now,
		}
		if err := uc.inviteLinks.Create(ctx, link); err != nil {
			return fmt.Errorf("create invite link group %d: %w", i, err)
		}

		group := &subscription.CorporateGroup{
			ID:              uuid.New().String(),
			Name:            fmt.Sprintf("Группа %d", i),
			PurchaseID:      event.PaymentID,
			ManagerID:       event.UserID,
			TeacherJoinCode: uuid.New().String(),
			StudentLinkID:   linkID,
			Semesters:       numSemesters,
			CreatedAt:       now,
		}
		if err := uc.corporateGroups.Create(ctx, group); err != nil {
			return fmt.Errorf("create group %d: %w", i, err)
		}
	}

	return uc.sendCorporatePDF(ctx, event.UserID, event.PaymentID)
}

// sendCorporatePDF generates and sends the corporate PDF to the manager via notification.
func (uc *UseCase) sendCorporatePDF(ctx context.Context, managerID int64, paymentID string) error {
	purchase, err := uc.corporateGroups.GetPurchaseByPaymentID(ctx, paymentID)
	if err != nil {
		_ = uc.notifications.Send(ctx, managerID, "Корпоративный пакет оформлен!")
		return nil
	}
	groups, err := uc.corporateGroups.GetGroupsByPurchaseID(ctx, paymentID)
	if err != nil {
		_ = uc.notifications.Send(ctx, managerID, "Корпоративный пакет оформлен!")
		return nil
	}

	pdfGroups := make([]pdfpkg.GroupInfo, len(groups))
	for i, g := range groups {
		pdfGroups[i] = pdfpkg.GroupInfo{
			Number:      i + 1,
			Name:        g.Name,
			TeacherLink: fmt.Sprintf("https://t.me/%s?start=corp_teacher_%s", uc.botUsername, g.TeacherJoinCode),
			StudentLink: fmt.Sprintf("https://t.me/%s?start=invite_%s", uc.botUsername, g.StudentLinkID),
		}
	}

	summary := pdfpkg.PlanSummary{
		Groups:      purchase.GroupsCount,
		Semesters:   purchase.Semesters,
		TotalAmount: purchase.TotalAmount,
	}
	data, genErr := pdfpkg.Generate(summary, pdfGroups)
	if genErr != nil {
		_ = uc.notifications.Send(ctx, managerID, "Корпоративный пакет оформлен! PDF формируется.")
		return nil
	}

	caption := fmt.Sprintf("Корпоративный пакет: %d групп × %d сем.", purchase.GroupsCount, purchase.Semesters)
	_ = uc.notifications.SendDocument(ctx, managerID, "mnemo_corporate.pdf", data, caption)
	return nil
}

