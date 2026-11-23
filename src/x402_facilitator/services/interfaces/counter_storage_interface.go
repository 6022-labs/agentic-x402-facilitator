package interfaces

import (
	"context"
	"time"
)

// CounterStorageInterface holds expiring counters every replica shares.
type CounterStorageInterface interface {
	// TryAdd adds amount to key unless the total would exceed limit; a new counter expires after ttl.
	TryAdd(ctx context.Context, key string, amount int64, limit int64, ttl time.Duration) (bool, error)
	// Subtract takes amount off key; a counter that expired stays gone.
	Subtract(ctx context.Context, key string, amount int64) error
	// Read returns 0 for a missing counter.
	Read(ctx context.Context, key string) (int64, error)
}
