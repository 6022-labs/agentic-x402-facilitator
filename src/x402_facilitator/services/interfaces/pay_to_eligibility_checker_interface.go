package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// PayToEligibilityCheckerInterface decides whether the facilitator settles payments to a payee.
type PayToEligibilityCheckerInterface interface {
	// Check returns a *common/errors.InvalidPaymentError for a payee it does not serve and an *UnavailableError
	// when the registry cannot be read.
	Check(ctx context.Context, payTo common.Address) error
}
