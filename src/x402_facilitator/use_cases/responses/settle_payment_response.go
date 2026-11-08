package responses

import (
	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
)

// SettlePaymentResponse is the x402 v2 SettlementResponse; Transaction is empty when nothing was broadcast.
type SettlePaymentResponse struct {
	Success      bool   `json:"success"`
	ErrorReason  string `json:"errorReason,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Payer        string `json:"payer,omitempty"`
	Transaction  string `json:"transaction"`
	Network      string `json:"network"`
}

func NewSettlePaymentResponseFromInvalidPayment(
	invalidPayment *apperrors.InvalidPaymentError,
	payer string,
	network string,
) *SettlePaymentResponse {
	return &SettlePaymentResponse{
		Success:      false,
		ErrorReason:  invalidPayment.Reason,
		ErrorMessage: invalidPayment.Message,
		Payer:        payer,
		Network:      network,
	}
}

func NewSettlePaymentResponseFromTransfer(
	transfer *models.Transfer,
	payer string,
	network string,
) *SettlePaymentResponse {
	response := &SettlePaymentResponse{
		Success:     transfer.Status == models.TransferStatusConfirmed,
		Payer:       payer,
		Transaction: transfer.TxHash.Hex(),
		Network:     network,
	}
	switch transfer.Status {
	case models.TransferStatusReverted:
		response.ErrorReason = enums.PaymentReasonInvalidTransactionState.String()
		response.ErrorMessage = "the settlement transaction reverted"
	case models.TransferStatusPending:
		response.ErrorReason = enums.PaymentReasonSettlementPending.String()
		response.ErrorMessage = "the settlement transaction is not confirmed yet"
	}
	return response
}

// Outcome labels the settlement for metrics: "settled" or the x402 reason code.
func (r *SettlePaymentResponse) Outcome() string {
	if r.Success {
		return "settled"
	}
	return r.ErrorReason
}
