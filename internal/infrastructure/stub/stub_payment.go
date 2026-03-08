package stub

import (
	"context"
	"fmt"

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

	switch plan {
	case "yearly":
		amount = 9900
	default:
		amount = 990
	}

	return invoiceID, paymentURL, amount, nil
}

// VerifyWebhookSignature always succeeds for the stub implementation.
func (s *PaymentService) VerifyWebhookSignature(_ []byte, _ string) error {
	return nil
}
