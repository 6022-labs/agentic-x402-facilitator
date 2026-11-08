package interfaces

import (
	"context"
	"math/big"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/ethereum/go-ethereum/common"
)

// Eip3009TokenRequesterInterface reads an EIP-3009 token's state without writing on-chain.
type Eip3009TokenRequesterInterface interface {
	// AuthorizationState reports whether the authorizer already used the nonce.
	AuthorizationState(
		ctx context.Context,
		chainId uint64,
		token common.Address,
		authorizer common.Address,
		nonce [32]byte,
	) (bool, error)
	BalanceOf(ctx context.Context, chainId uint64, token common.Address, owner common.Address) (*big.Int, error)
	// SimulateTransferWithAuthorization dry-runs the transfer from the relayer; a revert returns
	// common/errors.ErrTransferRejected, an unloadable relayer key common/errors.ErrRelayerUnavailable.
	SimulateTransferWithAuthorization(
		ctx context.Context,
		chainId uint64,
		token common.Address,
		authorization *models.TransferAuthorization,
	) error
}
