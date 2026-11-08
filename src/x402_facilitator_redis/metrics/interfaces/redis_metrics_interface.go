package interfaces

import (
	"context"
	"time"
)

// RedisMetricsInterface records db.client.* metrics for the shared redis store.
type RedisMetricsInterface interface {
	// RecordOperation records one redis command; operation is the redis verb (GET, SET, DEL, EVALSHA).
	RecordOperation(ctx context.Context, operation string, duration time.Duration, err error)
}
