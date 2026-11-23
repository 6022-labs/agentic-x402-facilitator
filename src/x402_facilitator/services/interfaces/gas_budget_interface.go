package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// GasBudgetInterface bounds the gas the relayer pays per chain and UTC day; a chain with no budget is unbounded.
type GasBudgetInterface interface {
	// Check returns an *common/errors.UnavailableError once the day's budget is spent or the store cannot be read.
	Check(ctx context.Context, chainId uint64) error
	// Reserve holds the transaction's worst-case cost in wei; common/errors.ErrGasBudgetSpent when it does not fit.
	Reserve(ctx context.Context, chainId uint64, txHash common.Hash, maxCost *big.Int) error
	// Settle replaces the transaction's reservation with the cost it paid in wei, zero for one never broadcast; only
	// the first call for a reservation counts.
	Settle(ctx context.Context, chainId uint64, txHash common.Hash, cost *big.Int)
}
