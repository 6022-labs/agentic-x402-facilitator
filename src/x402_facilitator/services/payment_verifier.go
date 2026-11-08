package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// validBeforeMargin leaves the settlement time to land before the authorization expires.
const validBeforeMargin = 6 * time.Second

type PaymentVerifier struct {
	tokensSettings          *x402_facilitator_settings.TokensSettings
	payToEligibilityChecker x402_facilitator_interfaces.PayToEligibilityCheckerInterface
	eip3009TokenRequester   x402_facilitator_interfaces.Eip3009TokenRequesterInterface
}

func NewPaymentVerifier(
	tokensSettings *x402_facilitator_settings.TokensSettings,
	payToEligibilityChecker x402_facilitator_interfaces.PayToEligibilityCheckerInterface,
	eip3009TokenRequester x402_facilitator_interfaces.Eip3009TokenRequesterInterface,
) *PaymentVerifier {
	return &PaymentVerifier{
		tokensSettings:          tokensSettings,
		payToEligibilityChecker: payToEligibilityChecker,
		eip3009TokenRequester:   eip3009TokenRequester,
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
	return v.checkSignature(payment, token)
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

func (v *PaymentVerifier) checkSignature(payment *models.Payment, token *x402_facilitator_settings.TokenConfig) error {
	authorization := payment.Authorization
	digest, err := helpers.TransferWithAuthorizationDigest(
		authorization,
		payment.Asset,
		token.Eip712Name,
		token.Eip712Version,
		payment.ChainId,
	)
	if err != nil {
		return apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidPayload, "authorization cannot be hashed")
	}

	signer, err := v.recoverSigner(digest, authorization.Signature)
	if err != nil || signer != authorization.From {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidSignature,
			"signature does not recover to from",
		)
	}
	return nil
}

// recoverSigner accepts a v of 27/28 or 0/1, as wallets emit either.
func (v *PaymentVerifier) recoverSigner(digest []byte, signature []byte) (common.Address, error) {
	normalized := make([]byte, len(signature))
	copy(normalized, signature)
	if normalized[64] >= 27 {
		normalized[64] -= 27
	}
	publicKey, err := crypto.SigToPub(digest, normalized)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*publicKey), nil
}

func (v *PaymentVerifier) checkOnChain(ctx context.Context, payment *models.Payment) error {
	authorization := payment.Authorization

	used, err := v.eip3009TokenRequester.AuthorizationState(
		ctx,
		payment.ChainId,
		payment.Asset,
		authorization.From,
		authorization.Nonce,
	)
	if err != nil {
		return v.unavailable(err)
	}
	if used {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidTransactionState,
			"authorization nonce is already used",
		)
	}

	balance, err := v.eip3009TokenRequester.BalanceOf(ctx, payment.ChainId, payment.Asset, authorization.From)
	if err != nil {
		return v.unavailable(err)
	}
	if balance.Cmp(authorization.Value) < 0 {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInsufficientFunds,
			"payer balance is below the amount",
		)
	}

	err = v.eip3009TokenRequester.SimulateTransferWithAuthorization(ctx, payment.ChainId, payment.Asset, authorization)
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
