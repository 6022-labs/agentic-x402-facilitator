package interfaces

import (
	"context"
	"time"
)

// CacheStorageInterface is the key-value store every replica shares.
type CacheStorageInterface interface {
	// Store sets key to value; a zero ttl keeps it until deleted.
	Store(ctx context.Context, key string, value string, ttl time.Duration) error
	// ReadString returns nil for a missing key.
	ReadString(ctx context.Context, key string) (*string, error)
	Delete(ctx context.Context, key string) error
}
