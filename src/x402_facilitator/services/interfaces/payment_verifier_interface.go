package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
)

// PaymentVerifierInterface runs every check a payment needs before it can settle, without writing on-chain.
type PaymentVerifierInterface interface {
	// Verify returns a *common/errors.InvalidPaymentError for a payment to refuse, an unserved payee or a payer at its
	// limit included, and an *UnavailableError when a dependency cannot be reached or the gas budget is spent.
	Verify(ctx context.Context, payment *models.Payment) error
	// VerifyAuthorization runs only the local network, token and signature checks: enough to trust that the request
	// carries the authorization it names. Returns a *common/errors.InvalidPaymentError.
	VerifyAuthorization(payment *models.Payment) error
}
