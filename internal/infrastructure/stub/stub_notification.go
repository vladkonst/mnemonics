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
