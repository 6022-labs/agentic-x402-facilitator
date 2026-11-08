package requests

import (
	"fmt"
	"math/big"
	"strings"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	signatureLength          = 65
	nonceLength              = 32
	extraAssetTransferMethod = "assetTransferMethod"
)

// PaymentRequest is the body of POST /verify and POST /settle.
type PaymentRequest struct {
	X402Version         int                  `json:"x402Version"`
	PaymentPayload      *PaymentPayload      `json:"paymentPayload"`
	PaymentRequirements *PaymentRequirements `json:"paymentRequirements"`
}

// ValidateAndSanitize returns a *apperrors.InvalidPaymentError naming the first structural defect.
func (r *PaymentRequest) ValidateAndSanitize() error {
	if r.PaymentPayload == nil {
		return apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidPayload, "paymentPayload is required")
	}
	if r.PaymentRequirements == nil {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"paymentRequirements is required",
		)
	}
	if r.X402Version != enums.X402Version || r.PaymentPayload.X402Version != enums.X402Version {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidX402Version,
			"only x402 version 2 is supported",
		)
	}

	requirements := r.PaymentRequirements
	if requirements.Scheme != enums.PaymentSchemeExact.String() {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonUnsupportedScheme,
			"only the exact scheme is supported",
		)
	}
	if r.PaymentPayload.AcceptedScheme() != requirements.Scheme {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidScheme,
			"payload scheme does not match the requirements",
		)
	}

	chainId, err := helpers.ParseCaip2(requirements.Network)
	if err != nil {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidNetwork,
			fmt.Sprintf("network %s", err.Error()),
		)
	}
	// One spelling per chain, since the network becomes a metric label.
	if requirements.Network != helpers.FormatCaip2(chainId) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidNetwork,
			"network must be written eip155:<chain id>",
		)
	}
	if r.PaymentPayload.AcceptedNetwork() != requirements.Network {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidNetwork,
			"payload network does not match the requirements",
		)
	}

	amount, err := r.validateRequirements()
	if err != nil {
		return err
	}
	if err := r.checkAccepted(); err != nil {
		return err
	}

	value, err := r.validateAuthorization()
	if err != nil {
		return err
	}
	if common.HexToAddress(r.PaymentPayload.Payload.Authorization.To) != common.HexToAddress(requirements.PayTo) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidRecipient,
			"authorization to does not match payTo",
		)
	}
	if value.Cmp(amount) != 0 {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidAuthorizationValue,
			"authorization value does not match amount",
		)
	}
	return nil
}

func (r *PaymentRequest) Network() string {
	if r.PaymentRequirements == nil {
		return ""
	}
	return r.PaymentRequirements.Network
}

// Payer is the authorization's from when it is an address, empty otherwise.
func (r *PaymentRequest) Payer() string {
	if r.PaymentPayload == nil || r.PaymentPayload.Payload.Authorization == nil {
		return ""
	}
	from := r.PaymentPayload.Payload.Authorization.From
	if !common.IsHexAddress(from) {
		return ""
	}
	return common.HexToAddress(from).Hex()
}

// validateRequirements returns the amount, to compare with the authorization value.
func (r *PaymentRequest) validateRequirements() (*big.Int, error) {
	requirements := r.PaymentRequirements

	if method, found := requirements.Extra[extraAssetTransferMethod]; found &&
		method != enums.AssetTransferMethodEip3009.String() {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"only the eip3009 asset transfer method is supported",
		)
	}
	if !common.IsHexAddress(requirements.Asset) {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"asset is not an address",
		)
	}
	if !common.IsHexAddress(requirements.PayTo) {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"payTo is not an address",
		)
	}
	amount, ok := r.parseUint(requirements.Amount)
	if !ok || amount.Sign() == 0 {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"amount must be a positive decimal integer",
		)
	}
	return amount, nil
}

// checkAccepted refuses a v2 payload signed for other terms than the submitted requirements.
func (r *PaymentRequest) checkAccepted() error {
	accepted := r.PaymentPayload.Accepted
	if accepted == nil {
		return nil
	}
	requirements := r.PaymentRequirements
	if accepted.Amount != requirements.Amount ||
		!strings.EqualFold(accepted.Asset, requirements.Asset) ||
		!strings.EqualFold(accepted.PayTo, requirements.PayTo) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"accepted requirements do not match the payment requirements",
		)
	}
	return nil
}

// validateAuthorization returns the value, to compare with the requirements amount.
func (r *PaymentRequest) validateAuthorization() (*big.Int, error) {
	raw := r.PaymentPayload.Payload.Authorization
	if raw == nil {
		return nil, apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidPayload, "authorization is required")
	}
	if !common.IsHexAddress(raw.From) || !common.IsHexAddress(raw.To) {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"authorization from and to must be addresses",
		)
	}

	value, valueOk := r.parseUint(raw.Value)
	_, validAfterOk := r.parseUint(raw.ValidAfter)
	_, validBeforeOk := r.parseUint(raw.ValidBefore)
	if !valueOk || !validAfterOk || !validBeforeOk {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"authorization value, validAfter and validBefore must be decimal integers",
		)
	}

	nonce, err := hexutil.Decode(raw.Nonce)
	if err != nil || len(nonce) != nonceLength {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"authorization nonce must be 32 bytes of hex",
		)
	}

	signature, err := hexutil.Decode(r.PaymentPayload.Payload.Signature)
	if err != nil || len(signature) != signatureLength {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidSignature,
			"signature must be 65 bytes of hex",
		)
	}
	return value, nil
}

// parseUint accepts a non-negative decimal integer that fits a uint256.
func (r *PaymentRequest) parseUint(s string) (*big.Int, bool) {
	value, ok := new(big.Int).SetString(strings.TrimSpace(s), 10)
	if !ok || value.Sign() < 0 || value.BitLen() > 256 {
		return nil, false
	}
	return value, true
}
