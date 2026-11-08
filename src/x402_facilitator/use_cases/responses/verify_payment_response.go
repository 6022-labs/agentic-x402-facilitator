package responses

import apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"

// VerifyPaymentResponse is the x402 v2 VerifyResponse.
type VerifyPaymentResponse struct {
	IsValid        bool   `json:"isValid"`
	InvalidReason  string `json:"invalidReason,omitempty"`
	InvalidMessage string `json:"invalidMessage,omitempty"`
	Payer          string `json:"payer,omitempty"`
}

func NewValidVerifyPaymentResponse(payer string) *VerifyPaymentResponse {
	return &VerifyPaymentResponse{IsValid: true, Payer: payer}
}

func NewVerifyPaymentResponseFromInvalidPayment(
	invalidPayment *apperrors.InvalidPaymentError,
	payer string,
) *VerifyPaymentResponse {
	return &VerifyPaymentResponse{
		IsValid:        false,
		InvalidReason:  invalidPayment.Reason,
		InvalidMessage: invalidPayment.Message,
		Payer:          payer,
	}
}

// Outcome labels the verification for metrics: "valid" or the x402 reason code.
func (r *VerifyPaymentResponse) Outcome() string {
	if r.IsValid {
		return "valid"
	}
	return r.InvalidReason
}
