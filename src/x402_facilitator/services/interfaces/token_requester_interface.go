package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TokenRequesterInterface reads a token's state without writing on-chain.
type TokenRequesterInterface interface {
	// AuthorizationState reports whether the authorizer already used the EIP-3009 nonce.
	AuthorizationState(
		ctx context.Context,
		chainId uint64,
		token common.Address,
		authorizer common.Address,
		nonce [32]byte,
	) (bool, error)
	BalanceOf(ctx context.Context, chainId uint64, token common.Address, owner common.Address) (*big.Int, error)
	Allowance(
		ctx context.Context,
		chainId uint64,
		token common.Address,
		owner common.Address,
		spender common.Address,
	) (*big.Int, error)
}
