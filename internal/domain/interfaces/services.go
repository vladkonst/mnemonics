package interfaces

import "context"

// ── External Service Contracts ────────────────────────────────────────────────

// StorageService abstracts S3-compatible object storage.
type StorageService interface {
	// PresignURL generates a temporary pre-signed download URL for an S3 key.
	PresignURL(ctx context.Context, s3Key string) (string, error)
}

// PaymentService abstracts the payment gateway integration.
type PaymentService interface {
	// CreateInvoice creates a payment invoice and returns the payment URL.
	CreateInvoice(ctx context.Context, userID int64, plan string) (invoiceID, paymentURL string, amount int, err error)
	// VerifyWebhookSignature validates the HMAC signature from the payment gateway.
	VerifyWebhookSignature(payload []byte, signature string) error
}

// NotificationService abstracts sending messages to Telegram users.
type NotificationService interface {
	// Send sends a text message to the specified Telegram user.
	Send(ctx context.Context, telegramID int64, message string) error
}
