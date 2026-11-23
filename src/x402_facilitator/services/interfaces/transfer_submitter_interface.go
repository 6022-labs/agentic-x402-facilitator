package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/ethereum/go-ethereum/common"
)

// TransferSubmitterInterface sends settlement calls from the relayer and follows their outcome.
type TransferSubmitterInterface interface {
	// Simulate dry-runs the call from the relayer; a revert returns common/errors.ErrTransferRejected, an
	// unloadable relayer key common/errors.ErrRelayerUnavailable.
	Simulate(ctx context.Context, chainId uint64, call *models.TransferCall) error
	// Broadcast returns ErrTransferRejected (chain would revert), ErrRelayerUnavailable (key unloadable) or
	// ErrGasBudgetSpent from common/errors; an error means nothing was broadcast, an ambiguous send that it was.
	Broadcast(ctx context.Context, chainId uint64, call *models.TransferCall) (common.Hash, error)
	// AwaitTransfer waits for the receipt and confirmations; a transfer still unconfirmed at the timeout is pending.
	AwaitTransfer(ctx context.Context, chainId uint64, txHash common.Hash) (*models.Transfer, error)
	// TransferStatus reads where a broadcast transfer stands; one the chain does not know yet is pending.
	TransferStatus(ctx context.Context, chainId uint64, txHash common.Hash) (*models.Transfer, error)
}
