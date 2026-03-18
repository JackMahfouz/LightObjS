package domain

import "context"

// Storage defines the interface for object operations, allowing
// for multiple providers (S3, GCS, Local) to be swapped easily.
type Storage interface {
	Upload(ctx context.Context, obj *Object, data []byte) error
	Download(ctx context.Context, key string) (*Object, []byte, error)
	Delete(ctx context.Context, key string) error
}
