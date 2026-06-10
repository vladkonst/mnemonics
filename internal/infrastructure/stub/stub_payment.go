package stub

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// PaymentService is a stub payment gateway that returns fake invoices.
type PaymentService struct{}

// NewPaymentService creates a new stub PaymentService.
func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// CreateInvoice returns a fake invoice with a placeholder payment URL.
func (s *PaymentService) CreateInvoice(_ context.Context, userID int64, plan string) (invoiceID, paymentURL string, amount int, err error) {
	invoiceID = uuid.NewString()
	paymentURL = fmt.Sprintf("https://stub-payment.example.com/pay/%s", invoiceID)

	switch {
	case plan == "yearly":
		amount = 2990
	case strings.HasPrefix(plan, "corporate"):
		// plan format: "corporate" or "corporate:groups:semesters"
		parts := strings.SplitN(plan, ":", 3)
		groups, semesters := 1, 1
		if len(parts) == 3 {
			if g, err := strconv.Atoi(parts[1]); err == nil && g > 0 {
				groups = g
			}
			if s, err := strconv.Atoi(parts[2]); err == nil && s > 0 {
				semesters = s
			}
		}
		amount = 4470 * groups * semesters
	default:
		amount = 149
	}

	return invoiceID, paymentURL, amount, nil
}

// VerifyWebhookSignature always succeeds for the stub implementation.
func (s *PaymentService) VerifyWebhookSignature(_ []byte, _ string) error {
	return nil
}
