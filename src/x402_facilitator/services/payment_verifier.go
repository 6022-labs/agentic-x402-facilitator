package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/samber/lo"
)

// validBeforeMargin leaves the settlement time to land before the authorization expires.
const validBeforeMargin = 6 * time.Second

// PaymentVerifier runs the checks every method shares and hands the method-specific ones to its strategy.
type PaymentVerifier struct {
	tokensSettings                *x402_facilitator_settings.TokensSettings
	payToEligibilityChecker       x402_facilitator_interfaces.PayToEligibilityCheckerInterface
	assetTransferMethodStrategies x402_facilitator_interfaces.AssetTransferMethodStrategiesInterface
	tokenRequester                x402_facilitator_interfaces.TokenRequesterInterface
	transferSubmitter             x402_facilitator_interfaces.TransferSubmitterInterface
}

func NewPaymentVerifier(
	tokensSettings *x402_facilitator_settings.TokensSettings,
	payToEligibilityChecker x402_facilitator_interfaces.PayToEligibilityCheckerInterface,
	assetTransferMethodStrategies x402_facilitator_interfaces.AssetTransferMethodStrategiesInterface,
	tokenRequester x402_facilitator_interfaces.TokenRequesterInterface,
	transferSubmitter x402_facilitator_interfaces.TransferSubmitterInterface,
) *PaymentVerifier {
	return &PaymentVerifier{
		tokensSettings:                tokensSettings,
		payToEligibilityChecker:       payToEligibilityChecker,
		assetTransferMethodStrategies: assetTransferMethodStrategies,
		tokenRequester:                tokenRequester,
		transferSubmitter:             transferSubmitter,
	}
}

func (v *PaymentVerifier) Verify(ctx context.Context, payment *models.Payment) error {
	// Local checks first, so a forged payment never costs a registry lookup.
	if err := v.VerifyAuthorization(payment); err != nil {
		return err
	}
	if err := v.checkValidityWindow(payment.Authorization); err != nil {
		return err
	}
	if err := v.payToEligibilityChecker.Check(ctx, payment.PayTo); err != nil {
		return err
	}
	return v.checkOnChain(ctx, payment)
}

func (v *PaymentVerifier) VerifyAuthorization(payment *models.Payment) error {
	if !v.tokensSettings.HasChain(payment.ChainId) {
		return apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidNetwork, "network is not supported")
	}
	token := v.tokensSettings.Get(payment.ChainId, payment.Asset)
	if token == nil {
		return apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidPaymentRequirements, "asset is not supported")
	}
	strategy, found := v.assetTransferMethodStrategies.For(payment.Method)
	if !found || !lo.Contains(token.AssetTransferMethods, payment.Method) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"asset is not settled with this transfer method",
		)
	}
	return strategy.VerifySignature(payment, token)
}

func (v *PaymentVerifier) checkValidityWindow(authorization *models.TransferAuthorization) error {
	now := time.Now()
	// Tokens require block.timestamp > validAfter, so an authorization valid from this second is not yet.
	if authorization.ValidAfter.Cmp(big.NewInt(now.Unix())) >= 0 {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidAuthorizationValidAfter,
			"authorization is not valid yet",
		)
	}
	if authorization.ValidBefore.Cmp(big.NewInt(now.Add(validBeforeMargin).Unix())) <= 0 {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidAuthorizationValidBefore,
			"authorization expires too soon",
		)
	}
	return nil
}

func (v *PaymentVerifier) checkOnChain(ctx context.Context, payment *models.Payment) error {
	strategy, _ := v.assetTransferMethodStrategies.For(payment.Method)
	if err := strategy.VerifyOnChain(ctx, payment); err != nil {
		return v.refusedOrUnavailable(err)
	}

	balance, err := v.tokenRequester.BalanceOf(ctx, payment.ChainId, payment.Asset, payment.Authorization.From)
	if err != nil {
		return v.unavailable(err)
	}
	if balance.Cmp(payment.Authorization.Value) < 0 {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInsufficientFunds,
			"payer balance is below the amount",
		)
	}

	call, err := strategy.TransferCall(payment)
	if err != nil {
		return err
	}
	err = v.transferSubmitter.Simulate(ctx, payment.ChainId, call)
	if errors.Is(err, apperrors.ErrTransferRejected) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidTransactionState,
			"transfer simulation reverts",
		)
	}
	if err != nil {
		return v.unavailable(err)
	}
	return nil
}

func (v *PaymentVerifier) refusedOrUnavailable(err error) error {
	var invalidPayment *apperrors.InvalidPaymentError
	if errors.As(err, &invalidPayment) {
		return err
	}
	return v.unavailable(err)
}

func (v *PaymentVerifier) unavailable(err error) error {
	if errors.Is(err, apperrors.ErrRelayerUnavailable) {
		return apperrors.NewUnavailableError("relayer_unavailable", "the relayer wallet is unavailable", err)
	}
	return apperrors.NewUnavailableError(
		"chain_unavailable",
		"the chain RPC is unavailable",
		fmt.Errorf("read token: %w", err),
	)
}
