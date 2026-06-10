package stub

import (
	"context"
	"fmt"
)

// NotificationService is a stub notification service that logs to stdout.
type NotificationService struct{}

// NewNotificationService creates a new stub NotificationService.
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// Send logs the notification message to stdout.
func (s *NotificationService) Send(_ context.Context, telegramID int64, message string) error {
	fmt.Printf("[NOTIFICATION] user=%d message=%q\n", telegramID, message)
	return nil
}

// SendDocument logs a document send to stdout (stub — does not actually send).
func (s *NotificationService) SendDocument(_ context.Context, telegramID int64, filename string, data []byte, caption string) error {
	fmt.Printf("[NOTIFICATION] user=%d document=%q size=%d caption=%q\n", telegramID, filename, len(data), caption)
	return nil
}
