package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
)

// PayerLimiterInterface caps the settlements of one payer per chain and hour.
type PayerLimiterInterface interface {
	// Check counts nothing. It returns a *common/errors.InvalidPaymentError once the cap is reached and an
	// *UnavailableError when the store cannot be read.
	Check(ctx context.Context, payment *models.Payment) error
	// Take counts one settlement and returns its hold, empty when no cap applies; it fails like Check.
	Take(ctx context.Context, payment *models.Payment) (string, error)
	// GiveBack uncounts a settlement that was not broadcast.
	GiveBack(ctx context.Context, hold string)
}
