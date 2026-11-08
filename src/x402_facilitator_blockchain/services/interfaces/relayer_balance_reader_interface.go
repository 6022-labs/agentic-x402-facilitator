package interfaces

import (
	"context"
	"math/big"
)

// RelayerBalanceReaderInterface reads the relayer's native balance on every configured chain.
type RelayerBalanceReaderInterface interface {
	// Balances maps each chain id to the relayer's balance in wei, leaving out chains that cannot be read.
	Balances(ctx context.Context) map[uint64]*big.Int
}
