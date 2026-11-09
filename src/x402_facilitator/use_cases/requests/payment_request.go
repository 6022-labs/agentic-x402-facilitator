package requests

import (
	"fmt"
	"strings"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/ethereum/go-ethereum/common"
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

	if err := r.validateRequirements(); err != nil {
		return err
	}
	return r.checkAccepted()
}

// AssetTransferMethod is how the payment moves the token; only meaningful after ValidateAndSanitize.
func (r *PaymentRequest) AssetTransferMethod() enums.AssetTransferMethod {
	method, _ := r.PaymentRequirements.AssetTransferMethod()
	return method
}

func (r *PaymentRequest) Network() string {
	if r.PaymentRequirements == nil {
		return ""
	}
	return r.PaymentRequirements.Network
}

// Payer is the signed authorization's from when it is an address, empty otherwise.
func (r *PaymentRequest) Payer() string {
	if r.PaymentPayload == nil {
		return ""
	}
	// Read from the method's own authorization: a payload may carry the other one to name someone else.
	method := enums.AssetTransferMethodEip3009
	if r.PaymentRequirements != nil {
		method, _ = r.PaymentRequirements.AssetTransferMethod()
	}
	from := ""
	switch {
	case method == enums.AssetTransferMethodPermit2 && r.PaymentPayload.Payload.Permit2Authorization != nil:
		from = r.PaymentPayload.Payload.Permit2Authorization.From
	case method == enums.AssetTransferMethodEip3009 && r.PaymentPayload.Payload.Eip3009Authorization != nil:
		from = r.PaymentPayload.Payload.Eip3009Authorization.From
	}
	if !common.IsHexAddress(from) {
		return ""
	}
	return common.HexToAddress(from).Hex()
}

func (r *PaymentRequest) validateRequirements() error {
	requirements := r.PaymentRequirements

	if _, known := requirements.AssetTransferMethod(); !known {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			fmt.Sprintf("assetTransferMethod must be one of %v", enums.GetAllAssetTransferMethods()),
		)
	}
	if !common.IsHexAddress(requirements.Asset) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"asset is not an address",
		)
	}
	if !common.IsHexAddress(requirements.PayTo) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"payTo is not an address",
		)
	}
	amount, ok := helpers.ParseUint256(requirements.Amount)
	if !ok || amount.Sign() == 0 {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"amount must be a positive decimal integer",
		)
	}
	return nil
}

// checkAccepted refuses a v2 payload signed for other terms than the submitted requirements.
func (r *PaymentRequest) checkAccepted() error {
	accepted := r.PaymentPayload.Accepted
	if accepted == nil {
		return nil
	}
	requirements := r.PaymentRequirements
	// The spec only asks clients to echo a non-default method, so an absent one is not a mismatch.
	_, echoed := accepted.Extra[extraAssetTransferMethod]
	acceptedMethod, _ := accepted.AssetTransferMethod()
	if (echoed && acceptedMethod != r.AssetTransferMethod()) ||
		accepted.Amount != requirements.Amount ||
		!strings.EqualFold(accepted.Asset, requirements.Asset) ||
		!strings.EqualFold(accepted.PayTo, requirements.PayTo) {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"accepted requirements do not match the payment requirements",
		)
	}
	return nil
}
