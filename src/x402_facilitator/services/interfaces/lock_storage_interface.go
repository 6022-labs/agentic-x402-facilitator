package interfaces

import (
	"context"
	"time"
)

// LockStorageInterface holds expiring locks every replica shares; only the owner's token may change one.
type LockStorageInterface interface {
	// TryLock takes key for ttl and returns its owner token, or false when another owner holds it.
	TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error)
	// Unlock frees key; common/errors.ErrLockNotOwned means token lost it to its expiry.
	Unlock(ctx context.Context, key string, token string) error
}
