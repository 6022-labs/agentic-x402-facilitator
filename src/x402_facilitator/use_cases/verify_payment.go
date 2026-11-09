package use_cases

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/metrics/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/responses"
)

type VerifyPayment struct {
	paymentReader
	paymentVerifier    x402_facilitator_interfaces.PaymentVerifierInterface
	settlementGuard    x402_facilitator_interfaces.SettlementGuardInterface
	facilitatorMetrics metrics_interfaces.FacilitatorMetricsInterface
}

func NewVerifyPayment(
	assetTransferMethodStrategies x402_facilitator_interfaces.AssetTransferMethodStrategiesInterface,
	paymentVerifier x402_facilitator_interfaces.PaymentVerifierInterface,
	settlementGuard x402_facilitator_interfaces.SettlementGuardInterface,
	facilitatorMetrics metrics_interfaces.FacilitatorMetricsInterface,
) *VerifyPayment {
	return &VerifyPayment{
		paymentReader:      paymentReader{assetTransferMethodStrategies: assetTransferMethodStrategies},
		paymentVerifier:    paymentVerifier,
		settlementGuard:    settlementGuard,
		facilitatorMetrics: facilitatorMetrics,
	}
}

func (u *VerifyPayment) Execute(
	ctx context.Context,
	request *requests.VerifyPaymentRequest,
) (*responses.VerifyPaymentResponse, error) {
	start := time.Now()

	var payment *models.Payment
	err := request.ValidateAndSanitize()
	if err == nil {
		payment, err = u.payment(&request.PaymentRequest)
	}
	response, err := u.invalid(err, request)
	if response == nil && err == nil {
		response, err = u.verify(ctx, request, payment)
	}

	outcome, verified := "", false
	if response != nil {
		outcome, verified = response.Outcome(), response.IsValid
	}
	u.facilitatorMetrics.RecordVerification(ctx, payment, verified, outcome, time.Since(start), err)

	return response, err
}

func (u *VerifyPayment) verify(
	ctx context.Context,
	request *requests.VerifyPaymentRequest,
	payment *models.Payment,
) (*responses.VerifyPaymentResponse, error) {
	if err := u.paymentVerifier.VerifyAuthorization(payment); err != nil {
		return u.invalid(err, request)
	}

	// Until it is mined, a payment already sent still verifies on-chain: a replay would get served.
	txHash, err := u.settlementGuard.RememberedTransaction(ctx, payment)
	var invalidPayment *apperrors.InvalidPaymentError
	if errors.As(err, &invalidPayment) {
		return u.invalid(err, request)
	}
	if err != nil {
		return nil, apperrors.NewUnavailableError(
			"store_unavailable",
			"the settlement store is unavailable",
			fmt.Errorf("settlement store: %w", err),
		)
	}
	if txHash != nil {
		return u.invalid(apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidTransactionState,
			"this authorization was already sent for settlement",
		), request)
	}

	if err := u.paymentVerifier.Verify(ctx, payment); err != nil {
		return u.invalid(err, request)
	}
	return responses.NewValidVerifyPaymentResponse(request.Payer()), nil
}

// invalid answers an invalid payment as a refusal; nil, nil when err is nil.
func (u *VerifyPayment) invalid(
	err error,
	request *requests.VerifyPaymentRequest,
) (*responses.VerifyPaymentResponse, error) {
	if err == nil {
		return nil, nil
	}
	var invalidPayment *apperrors.InvalidPaymentError
	if errors.As(err, &invalidPayment) {
		return responses.NewVerifyPaymentResponseFromInvalidPayment(invalidPayment, request.Payer()), nil
	}
	return nil, err
}
