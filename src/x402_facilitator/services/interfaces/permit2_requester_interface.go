package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Permit2RequesterInterface reads Permit2 and the x402 Permit2 proxy without writing on-chain.
type Permit2RequesterInterface interface {
	// NonceUsed reports whether the owner already used the Permit2 nonce, on any token.
	NonceUsed(ctx context.Context, chainId uint64, owner common.Address, nonce *big.Int) (bool, error)
	// ProxyDeployed reports whether x402ExactPermit2Proxy has code: a call to an empty address succeeds silently.
	ProxyDeployed(ctx context.Context, chainId uint64) (bool, error)
}
