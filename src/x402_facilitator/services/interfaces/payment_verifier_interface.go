package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
)

// PaymentVerifierInterface runs every check a payment needs before it can settle, without writing on-chain.
type PaymentVerifierInterface interface {
	// Verify returns a *common/errors.InvalidPaymentError for a payment to refuse, including a payee it does not
	// serve, and an *UnavailableError when the chain, the registry or the relayer cannot be reached.
	Verify(ctx context.Context, payment *models.Payment) error
	// VerifyAuthorization runs only the local network, token and signature checks: enough to trust that the request
	// carries the authorization it names. Returns a *common/errors.InvalidPaymentError.
	VerifyAuthorization(payment *models.Payment) error
}
