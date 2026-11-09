package services

import (
	"context"
	"fmt"
	"strings"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

const transferWithAuthorizationAbi = `[{"type":"function","name":"transferWithAuthorization","inputs":[
{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"value","type":"uint256"},
{"name":"validAfter","type":"uint256"},{"name":"validBefore","type":"uint256"},{"name":"nonce","type":"bytes32"},
{"name":"v","type":"uint8"},{"name":"r","type":"bytes32"},{"name":"s","type":"bytes32"}],"outputs":[]}]`

// Eip3009AssetTransferMethodStrategy settles with the token's own transferWithAuthorization(v, r, s).
type Eip3009AssetTransferMethodStrategy struct {
	paymentRequirementsMapper
	tokenRequester x402_facilitator_interfaces.TokenRequesterInterface
	tokenAbi       abi.ABI
}

func NewEip3009AssetTransferMethodStrategy(
	logger *zap.Logger,
	tokenRequester x402_facilitator_interfaces.TokenRequesterInterface,
) *Eip3009AssetTransferMethodStrategy {
	tokenAbi, err := abi.JSON(strings.NewReader(transferWithAuthorizationAbi))
	if err != nil {
		logger.Fatal("failed to parse the transferWithAuthorization ABI", zap.Error(err))
	}
	return &Eip3009AssetTransferMethodStrategy{tokenRequester: tokenRequester, tokenAbi: tokenAbi}
}

func (s *Eip3009AssetTransferMethodStrategy) Method() enums.AssetTransferMethod {
	return enums.AssetTransferMethodEip3009
}

func (s *Eip3009AssetTransferMethodStrategy) Payment(request *requests.PaymentRequest) (*models.Payment, error) {
	raw := request.PaymentPayload.Payload.Eip3009Authorization
	if raw == nil {
		return nil, apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidPayload, "authorization is required")
	}
	if !common.IsHexAddress(raw.From) || !common.IsHexAddress(raw.To) {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"authorization from and to must be addresses",
		)
	}
	value, valueOk := helpers.ParseUint256(raw.Value)
	validAfter, validAfterOk := helpers.ParseUint256(raw.ValidAfter)
	validBefore, validBeforeOk := helpers.ParseUint256(raw.ValidBefore)
	if !valueOk || !validAfterOk || !validBeforeOk {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"authorization value, validAfter and validBefore must be decimal integers",
		)
	}
	nonce, err := hexutil.Decode(raw.Nonce)
	if err != nil || len(nonce) != 32 {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"authorization nonce must be 32 bytes of hex",
		)
	}
	signature, err := hexutil.Decode(request.PaymentPayload.Payload.Signature)
	if err != nil || len(signature) != 65 {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidSignature,
			"signature must be 65 bytes of hex",
		)
	}

	payment := s.payment(request, enums.AssetTransferMethodEip3009)
	if common.HexToAddress(raw.To) != payment.PayTo {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidRecipient,
			"authorization to does not match payTo",
		)
	}
	if value.Cmp(payment.Amount) != 0 {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidAuthorizationValue,
			"authorization value does not match amount",
		)
	}

	payment.Authorization = &models.TransferAuthorization{
		From:        common.HexToAddress(raw.From),
		To:          payment.PayTo,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Signature:   signature,
	}
	copy(payment.Authorization.Nonce[:], nonce)
	return payment, nil
}

func (s *Eip3009AssetTransferMethodStrategy) VerifySignature(
	payment *models.Payment,
	token *x402_facilitator_settings.TokenConfig,
) error {
	digest, err := helpers.TransferWithAuthorizationDigest(
		payment.Authorization,
		payment.Asset,
		token.Eip712Name,
		token.Eip712Version,
		payment.ChainId,
	)
	if err != nil {
		return apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidPayload, "authorization cannot be hashed")
	}
	signer, err := helpers.RecoverSigner(digest, payment.Authorization.Signature)
	if err != nil || signer != payment.Authorization.From {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidSignature,
			"signature does not recover to from",
		)
	}
	return nil
}

func (s *Eip3009AssetTransferMethodStrategy) VerifyOnChain(ctx context.Context, payment *models.Payment) error {
	used, err := s.tokenRequester.AuthorizationState(
		ctx,
		payment.ChainId,
		payment.Asset,
		payment.Authorization.From,
		payment.Authorization.Nonce,
	)
	if err != nil {
		return err
	}
	if used {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidTransactionState,
			"authorization nonce is already used",
		)
	}
	return nil
}

func (s *Eip3009AssetTransferMethodStrategy) TransferCall(payment *models.Payment) (*models.TransferCall, error) {
	authorization := payment.Authorization
	v, r, sigS := helpers.SplitSignature(authorization.Signature)
	data, err := s.tokenAbi.Pack(
		"transferWithAuthorization",
		authorization.From,
		authorization.To,
		authorization.Value,
		authorization.ValidAfter,
		authorization.ValidBefore,
		authorization.Nonce,
		v,
		r,
		sigS,
	)
	if err != nil {
		return nil, fmt.Errorf("pack transferWithAuthorization: %w", err)
	}
	return &models.TransferCall{To: payment.Asset, Data: data}, nil
}
