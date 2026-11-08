package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/ethereum/go-ethereum/common"
)

// SettlementGuardInterface keeps one settlement per authorization in flight across replicas and remembers its
// transaction, so a retry reports it instead of sending again.
type SettlementGuardInterface interface {
	// TryAcquire returns the hold's token, or false while another settlement holds key.
	TryAcquire(ctx context.Context, key string) (string, bool, error)
	Release(ctx context.Context, key string, token string)
	// RememberTransaction keeps txHash with the payment's terms until the authorization expires, at least the lock
	// ttl, at most an hour.
	RememberTransaction(ctx context.Context, payment *models.Payment, txHash common.Hash) error
	// RememberedTransaction returns the transaction sent for the payment, nil when there is none, and a
	// *common/errors.InvalidPaymentError when it was sent for other terms under the same nonce.
	RememberedTransaction(ctx context.Context, payment *models.Payment) (*common.Hash, error)
}
