package use_cases

import (
	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
)

// paymentReader reads a validated request's payment through its asset transfer method's strategy.
type paymentReader struct {
	assetTransferMethodStrategies x402_facilitator_interfaces.AssetTransferMethodStrategiesInterface
}

func (r paymentReader) payment(request *requests.PaymentRequest) (*models.Payment, error) {
	strategy, found := r.assetTransferMethodStrategies.For(request.AssetTransferMethod())
	if !found {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"asset transfer method is not settled",
		)
	}
	return strategy.Payment(request)
}
