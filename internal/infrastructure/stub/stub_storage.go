// Package stub provides no-op / placeholder implementations of external service interfaces.
package stub

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// StorageService is a stub S3 storage service that saves files locally.
type StorageService struct {
	uploadsDir string
}

// NewStorageService creates a new stub StorageService that saves files to uploadsDir.
func NewStorageService(uploadsDir string) *StorageService {
	return &StorageService{uploadsDir: uploadsDir}
}

// UploadFile saves the file locally to the uploads directory.
func (s *StorageService) UploadFile(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	if err := os.MkdirAll(s.uploadsDir, 0755); err != nil {
		return err
	}
	dst, err := os.Create(filepath.Join(s.uploadsDir, key))
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, body)
	return err
}

// PresignURL returns a placeholder local URL for the given key.
func (s *StorageService) PresignURL(_ context.Context, s3Key string) (string, error) {
	return fmt.Sprintf("/api/v1/uploads/%s", s3Key), nil
}
