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

type SettlePayment struct {
	paymentReader
	paymentVerifier    x402_facilitator_interfaces.PaymentVerifierInterface
	settlementGuard    x402_facilitator_interfaces.SettlementGuardInterface
	transferSubmitter  x402_facilitator_interfaces.TransferSubmitterInterface
	facilitatorMetrics metrics_interfaces.FacilitatorMetricsInterface
}

func NewSettlePayment(
	assetTransferMethodStrategies x402_facilitator_interfaces.AssetTransferMethodStrategiesInterface,
	paymentVerifier x402_facilitator_interfaces.PaymentVerifierInterface,
	settlementGuard x402_facilitator_interfaces.SettlementGuardInterface,
	transferSubmitter x402_facilitator_interfaces.TransferSubmitterInterface,
	facilitatorMetrics metrics_interfaces.FacilitatorMetricsInterface,
) *SettlePayment {
	return &SettlePayment{
		paymentReader:      paymentReader{assetTransferMethodStrategies: assetTransferMethodStrategies},
		paymentVerifier:    paymentVerifier,
		settlementGuard:    settlementGuard,
		transferSubmitter:  transferSubmitter,
		facilitatorMetrics: facilitatorMetrics,
	}
}

func (u *SettlePayment) Execute(
	ctx context.Context,
	request *requests.SettlePaymentRequest,
) (*responses.SettlePaymentResponse, error) {
	start := time.Now()

	var payment *models.Payment
	verified := false
	err := request.ValidateAndSanitize()
	if err == nil {
		payment, err = u.payment(&request.PaymentRequest)
	}
	response, err := u.invalid(err, request)
	if response == nil && err == nil {
		response, verified, err = u.settle(ctx, request, payment)
	}

	outcome := ""
	if response != nil {
		outcome = response.Outcome()
	}
	u.facilitatorMetrics.RecordSettlement(ctx, payment, verified, outcome, time.Since(start), err)

	return response, err
}

// settle reports verified once the payment passed verification, now or for the transaction already sent.
func (u *SettlePayment) settle(
	ctx context.Context,
	request *requests.SettlePaymentRequest,
	payment *models.Payment,
) (*responses.SettlePaymentResponse, bool, error) {
	// Only the payer can then read the transaction remembered for its authorization.
	if err := u.paymentVerifier.VerifyAuthorization(payment); err != nil {
		response, err := u.invalid(err, request)
		return response, false, err
	}

	// A retry reports the transaction already sent: verifying again would refuse its used nonce.
	response, err := u.reportRemembered(ctx, request, payment)
	if err != nil {
		response, err := u.invalid(err, request)
		return response, false, err
	}
	if response != nil {
		return response, true, nil
	}

	if err := u.paymentVerifier.Verify(ctx, payment); err != nil {
		response, err := u.invalid(err, request)
		return response, false, err
	}

	response, err = u.settleVerified(ctx, request, payment)
	return response, true, err
}

func (u *SettlePayment) settleVerified(
	ctx context.Context,
	request *requests.SettlePaymentRequest,
	payment *models.Payment,
) (*responses.SettlePaymentResponse, error) {
	key := payment.SettlementKey()
	token, acquired, err := u.settlementGuard.TryAcquire(ctx, key)
	if err != nil {
		return nil, u.storeUnavailable(err)
	}
	if !acquired {
		return nil, apperrors.NewUnavailableError(
			"settlement_in_progress",
			"a settlement of this authorization is in progress, retry later",
			nil,
		)
	}

	// Another replica may have sent it between the first check and the lock.
	if response, err := u.reportRemembered(ctx, request, payment); response != nil || err != nil {
		u.settlementGuard.Release(ctx, key, token)
		if err != nil {
			return u.invalid(err, request)
		}
		return response, nil
	}

	strategy, _ := u.assetTransferMethodStrategies.For(payment.Method)
	call, err := strategy.TransferCall(payment)
	if err != nil {
		u.settlementGuard.Release(ctx, key, token)
		return nil, err
	}
	txHash, err := u.transferSubmitter.Broadcast(ctx, payment.ChainId, call)
	if err != nil {
		u.settlementGuard.Release(ctx, key, token)
		return u.broadcastFailure(err, request)
	}
	// Unremembered, the lock stays held until it expires, so no retry sends the authorization again.
	if err := u.settlementGuard.RememberTransaction(ctx, payment, txHash); err == nil {
		u.settlementGuard.Release(ctx, key, token)
	}

	transfer, err := u.transferSubmitter.AwaitTransfer(ctx, payment.ChainId, txHash)
	if err != nil {
		return nil, apperrors.NewUnavailableError(
			"chain_unavailable",
			"the chain RPC is unavailable",
			fmt.Errorf("await transfer: %w", err),
		)
	}
	return responses.NewSettlePaymentResponseFromTransfer(transfer, request.Payer(), request.Network()), nil
}

// reportRemembered answers from the transaction already sent; nil, nil when there is none.
func (u *SettlePayment) reportRemembered(
	ctx context.Context,
	request *requests.SettlePaymentRequest,
	payment *models.Payment,
) (*responses.SettlePaymentResponse, error) {
	txHash, err := u.settlementGuard.RememberedTransaction(ctx, payment)
	var invalidPayment *apperrors.InvalidPaymentError
	if errors.As(err, &invalidPayment) {
		return nil, err
	}
	if err != nil {
		return nil, u.storeUnavailable(err)
	}
	if txHash == nil {
		return nil, nil
	}

	transfer, err := u.transferSubmitter.TransferStatus(ctx, payment.ChainId, *txHash)
	if err != nil {
		return nil, apperrors.NewUnavailableError(
			"chain_unavailable",
			"the chain RPC is unavailable",
			fmt.Errorf("read transfer status: %w", err),
		)
	}
	// Never success: answering a settled authorization as paid again would let a replayed payment through.
	if transfer.Status == models.TransferStatusConfirmed {
		response := responses.NewSettlePaymentResponseFromInvalidPayment(
			apperrors.NewInvalidPaymentError(
				enums.PaymentReasonInvalidTransactionState,
				"this authorization was already settled in this transaction",
			),
			request.Payer(),
			request.Network(),
		)
		response.Transaction = transfer.TxHash.Hex()
		return response, nil
	}
	return responses.NewSettlePaymentResponseFromTransfer(transfer, request.Payer(), request.Network()), nil
}

func (u *SettlePayment) broadcastFailure(
	err error,
	request *requests.SettlePaymentRequest,
) (*responses.SettlePaymentResponse, error) {
	switch {
	case errors.Is(err, apperrors.ErrTransferRejected):
		return u.invalid(apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidTransactionState,
			"the chain rejects the transfer",
		), request)
	case errors.Is(err, apperrors.ErrRelayerUnavailable):
		return nil, apperrors.NewUnavailableError("relayer_unavailable", "the relayer wallet is unavailable", err)
	default:
		return nil, apperrors.NewUnavailableError(
			"settlement_unavailable",
			"the settlement could not be submitted",
			fmt.Errorf("broadcast transfer: %w", err),
		)
	}
}

func (u *SettlePayment) storeUnavailable(err error) error {
	return apperrors.NewUnavailableError(
		"store_unavailable",
		"the settlement store is unavailable",
		fmt.Errorf("settlement store: %w", err),
	)
}

// invalid answers an invalid payment as a refusal; nil, nil when err is nil.
func (u *SettlePayment) invalid(
	err error,
	request *requests.SettlePaymentRequest,
) (*responses.SettlePaymentResponse, error) {
	if err == nil {
		return nil, nil
	}
	var invalidPayment *apperrors.InvalidPaymentError
	if errors.As(err, &invalidPayment) {
		return responses.NewSettlePaymentResponseFromInvalidPayment(
			invalidPayment,
			request.Payer(),
			request.Network(),
		), nil
	}
	return nil, err
}
