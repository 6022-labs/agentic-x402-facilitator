package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
)

// AssetTransferMethodStrategyInterface owns everything one asset transfer method does differently.
type AssetTransferMethodStrategyInterface interface {
	Method() enums.AssetTransferMethod
	// Payment reads the method's signed authorization from a request that passed ValidateAndSanitize; a
	// *common/errors.InvalidPaymentError when it is malformed or signed for other requirements.
	Payment(request *requests.PaymentRequest) (*models.Payment, error)
	// VerifySignature checks locally that the payer signed the payment; a *common/errors.InvalidPaymentError if not.
	VerifySignature(payment *models.Payment, token *x402_facilitator_settings.TokenConfig) error
	// VerifyOnChain checks the authorization is unused and can pull the tokens; a *common/errors.InvalidPaymentError
	// for a payment to refuse, a plain error when the chain cannot be read.
	VerifyOnChain(ctx context.Context, payment *models.Payment) error
	// TransferCall is the relayer transaction that settles the payment.
	TransferCall(payment *models.Payment) (*models.TransferCall, error)
}
