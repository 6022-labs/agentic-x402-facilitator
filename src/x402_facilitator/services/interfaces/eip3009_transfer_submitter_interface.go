package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/ethereum/go-ethereum/common"
)

// Eip3009TransferSubmitterInterface submits transferWithAuthorization from the relayer and follows its outcome.
type Eip3009TransferSubmitterInterface interface {
	// Broadcast returns ErrTransferRejected (chain would revert) or ErrRelayerUnavailable (key unloadable) from
	// common/errors; an error means nothing was broadcast, and an ambiguous send counts as broadcast.
	Broadcast(
		ctx context.Context,
		chainId uint64,
		token common.Address,
		authorization *models.TransferAuthorization,
	) (common.Hash, error)
	// AwaitTransfer waits for the receipt and confirmations; a transfer still unconfirmed at the timeout is pending.
	AwaitTransfer(ctx context.Context, chainId uint64, txHash common.Hash) (*models.Transfer, error)
	// TransferStatus reads where a broadcast transfer stands; one the chain does not know yet is pending.
	TransferStatus(ctx context.Context, chainId uint64, txHash common.Hash) (*models.Transfer, error)
}
