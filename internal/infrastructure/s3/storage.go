// Package s3 provides an S3-compatible storage service implementation using aws-sdk-go-v2.
package s3

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// StorageService implements interfaces.StorageService backed by an S3-compatible store.
type StorageService struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	presignExpiry time.Duration
}

// Config holds the S3 connection parameters.
type Config struct {
	Endpoint      string
	Bucket        string
	Region        string
	AccessKey     string
	SecretKey     string
	PresignExpiry time.Duration
}

// NewStorageService creates a new S3 StorageService.
func NewStorageService(cfg Config) *StorageService {
	expiry := cfg.PresignExpiry
	if expiry == 0 {
		expiry = 15 * time.Minute
	}

	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: cfg.Endpoint}, nil
			},
		),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &StorageService{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
		presignExpiry: expiry,
	}
}

// UploadFile uploads a file to S3 under the given key.
func (s *StorageService) UploadFile(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	return err
}

// PresignURL generates a pre-signed GET URL for the given S3 key.
func (s *StorageService) PresignURL(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.presignExpiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
