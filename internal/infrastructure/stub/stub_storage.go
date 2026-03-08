// Package stub provides no-op / placeholder implementations of external service interfaces.
package stub

import (
	"context"
	"fmt"
)

// StorageService is a stub S3 storage service that returns placeholder URLs.
type StorageService struct{}

// NewStorageService creates a new stub StorageService.
func NewStorageService() *StorageService {
	return &StorageService{}
}

// PresignURL returns a placeholder URL for the given S3 key.
func (s *StorageService) PresignURL(_ context.Context, s3Key string) (string, error) {
	return fmt.Sprintf("https://stub-storage.example.com/%s", s3Key), nil
}
