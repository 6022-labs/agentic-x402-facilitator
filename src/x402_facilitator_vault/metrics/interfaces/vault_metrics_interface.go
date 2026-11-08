package interfaces

import (
	"context"
	"time"
)

// VaultMetricsInterface records Vault secret-operation health.
type VaultMetricsInterface interface {
	RecordOperation(ctx context.Context, operation, mount string, duration time.Duration, err error)
}
