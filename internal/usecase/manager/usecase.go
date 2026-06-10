// Package manager provides use cases for corporate purchase managers.
package manager

import (
	"context"
	"fmt"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/infrastructure/pdf"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// UseCase orchestrates manager operations (corporate purchase history, PDF regeneration).
type UseCase struct {
	corporateGroups interfaces.CorporateGroupRepository
	pdf             pdf.GenerateFunc
	botUsername     string
}

// NewUseCase creates a new manager UseCase.
func NewUseCase(
	corporateGroups interfaces.CorporateGroupRepository,
	pdfFn pdf.GenerateFunc,
	botUsername string,
) *UseCase {
	return &UseCase{
		corporateGroups: corporateGroups,
		pdf:             pdfFn,
		botUsername:     botUsername,
	}
}

// GetPurchases returns all corporate purchases made by a manager.
func (uc *UseCase) GetPurchases(ctx context.Context, managerID int64) ([]*subscription.CorporatePurchase, error) {
	return uc.corporateGroups.GetPurchasesByManagerID(ctx, managerID)
}

// GetAllPurchases returns all corporate purchases across all managers (admin view).
func (uc *UseCase) GetAllPurchases(ctx context.Context) ([]*subscription.CorporatePurchase, error) {
	return uc.corporateGroups.GetAllPurchases(ctx)
}

// GetPurchaseByID returns a single purchase by its payment_id (admin view, no owner check).
func (uc *UseCase) GetPurchaseByID(ctx context.Context, paymentID string) (*subscription.CorporatePurchaseWithGroups, error) {
	purchase, err := uc.corporateGroups.GetPurchaseByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	groups, err := uc.corporateGroups.GetGroupsByPurchaseID(ctx, purchase.PaymentID)
	if err != nil {
		return nil, err
	}
	return &subscription.CorporatePurchaseWithGroups{CorporatePurchase: purchase, Groups: groups}, nil
}

// GetAllGroups returns all corporate groups (admin view).
func (uc *UseCase) GetAllGroups(ctx context.Context) ([]*subscription.CorporateGroup, error) {
	return uc.corporateGroups.GetAllGroups(ctx)
}

// GetGroupsByPurchase returns all groups for a given purchase.
func (uc *UseCase) GetGroupsByPurchase(ctx context.Context, purchaseID string) ([]*subscription.CorporateGroup, error) {
	return uc.corporateGroups.GetGroupsByPurchaseID(ctx, purchaseID)
}

// GetGroupsByManager returns all groups for a given manager.
func (uc *UseCase) GetGroupsByManager(ctx context.Context, managerID int64) ([]*subscription.CorporateGroup, error) {
	return uc.corporateGroups.GetByManagerID(ctx, managerID)
}

// GetGroupByID returns a single corporate group by its ID.
func (uc *UseCase) GetGroupByID(ctx context.Context, id string) (*subscription.CorporateGroup, error) {
	return uc.corporateGroups.GetByID(ctx, id)
}

// GeneratePurchasePDF regenerates the corporate PDF for a past purchase.
// Pass managerID=0 to skip the owner check (admin access).
func (uc *UseCase) GeneratePurchasePDF(ctx context.Context, managerID int64, paymentID string) ([]byte, error) {
	purchase, err := uc.corporateGroups.GetPurchaseByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if managerID != 0 && purchase.ManagerID != managerID {
		return nil, apperrors.ErrForbidden
	}
	groups, err := uc.corporateGroups.GetGroupsByPurchaseID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	pdfGroups := make([]pdf.GroupInfo, len(groups))
	for i, g := range groups {
		pdfGroups[i] = pdf.GroupInfo{
			Number:      i + 1,
			Name:        g.Name,
			TeacherLink: fmt.Sprintf("https://t.me/%s?start=corp_teacher_%s", uc.botUsername, g.TeacherJoinCode),
			StudentLink: fmt.Sprintf("https://t.me/%s?start=invite_%s", uc.botUsername, g.StudentLinkID),
		}
	}

	summary := pdf.PlanSummary{
		Groups:      purchase.GroupsCount,
		Semesters:   purchase.Semesters,
		TotalAmount: purchase.TotalAmount,
	}
	return uc.pdf(summary, pdfGroups)
}
