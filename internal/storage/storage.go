package storage

import (
	"context"
	"io"
)

// Provider is the single interface every storage backend must satisfy.
// To add a new backend (S3, R2, GCS…) just implement this interface.
type Provider interface {
	// Upload stores the content of r under the given key.
	// Returns the public URL that can be served to clients.
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (url string, err error)

	// Delete removes the object at key. Returns nil if the object did not exist.
	Delete(ctx context.Context, key string) error

	// PublicURL returns the publicly accessible URL for a stored key,
	// without performing any I/O. Useful for building URLs from DB-stored keys.
	PublicURL(key string) string
}
