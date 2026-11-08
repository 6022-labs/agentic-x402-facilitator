package interfaces

import (
	"context"
	"math/big"
)

// RelayerMetricsInterface records the gas the relayer's settlements cost.
type RelayerMetricsInterface interface {
	// RecordGasSpent records one mined settlement's gas; status is "confirmed" or "reverted".
	RecordGasSpent(ctx context.Context, chainId uint64, gasUsed uint64, costWei *big.Int, status string)
}
