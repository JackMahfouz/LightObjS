package storage

import "io"

type Storage interface {
	Get(bucket, key string) (io.ReadCloser, error)
	Put(bucket, key string, data io.Reader) error
	Delete(bucket, key string) error
}
