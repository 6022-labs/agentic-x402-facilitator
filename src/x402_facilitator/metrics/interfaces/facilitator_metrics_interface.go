package interfaces

import (
	"context"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
)

// FacilitatorMetricsInterface records x402 verifications and settlements per network, token and verified payee.
type FacilitatorMetricsInterface interface {
	// RecordVerification records one /verify: payment is nil when the request never parsed, verified once it passed
	// verification, outcome is "valid" or an x402 reason code, err a failure that has none.
	RecordVerification(
		ctx context.Context,
		payment *models.Payment,
		verified bool,
		outcome string,
		duration time.Duration,
		err error,
	)
	// RecordSettlement records one /settle: verified once the payment passed verification, outcome is "settled" or
	// an x402 reason code, err a failure that has none.
	RecordSettlement(
		ctx context.Context,
		payment *models.Payment,
		verified bool,
		outcome string,
		duration time.Duration,
		err error,
	)
}
