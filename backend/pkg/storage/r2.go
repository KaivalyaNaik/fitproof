package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Service struct {
	client    *s3.Client
	bucket    string
	publicURL string // e.g. https://pub-xxx.r2.dev  (no trailing slash)
}

func NewR2(accountID, accessKeyID, secretAccessKey, bucket, publicURL string) (*R2Service, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID, secretAccessKey, "",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("load r2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2Service{client: client, bucket: bucket, publicURL: publicURL}, nil
}

// Upload streams r to R2 and returns the object key (used as the file ID).
func (s *R2Service) Upload(ctx context.Context, name, contentType string, r io.Reader) (string, error) {
	key := name
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("r2 put object: %w", err)
	}
	return key, nil
}

// Delete removes an object from R2 by its key.
func (s *R2Service) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("r2 delete object %s: %w", key, err)
	}
	return nil
}

// PublicURL returns the public URL for a given object key.
func (s *R2Service) PublicURL(key string) string {
	if s.publicURL == "" {
		return ""
	}
	return s.publicURL + "/" + key
}
