package storage

import (
	"LightObjS/internal/domain"
	"context"
	"fmt"
)

// S3Provider implements the domain.Storage interface for AWS S3.
type S3Provider struct {
	Bucket string
	Region string
}

func NewS3Provider(bucket, region string) *S3Provider {
	return &S3Provider{
		Bucket: bucket,
		Region: region,
	}
}

func (s *S3Provider) Upload(ctx context.Context, obj *domain.Object, data []byte) error {
	fmt.Printf("[S3] Uploading %s to bucket %s (Size: %d bytes)\n", obj.Key, s.Bucket, obj.Size)
	// Integration with AWS SDK would happen here
	return nil
}

func (s *S3Provider) Download(ctx context.Context, key string) (*domain.Object, []byte, error) {
	fmt.Printf("[S3] Downloading %s from bucket %s\n", key, s.Bucket)
	return &domain.Object{Key: key}, []byte("mock-data"), nil
}

func (s *S3Provider) Delete(ctx context.Context, key string) error {
	fmt.Printf("[S3] Deleting %s from bucket %s\n", key, s.Bucket)
	return nil
}
