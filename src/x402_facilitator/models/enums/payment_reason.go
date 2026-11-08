package enums

// PaymentReason is an x402 v2 verification or settlement reason code (specification §9).
type PaymentReason string

const (
	PaymentReasonInsufficientFunds               PaymentReason = "insufficient_funds"
	PaymentReasonInvalidAuthorizationValidAfter  PaymentReason = "invalid_exact_evm_payload_authorization_valid_after"
	PaymentReasonInvalidAuthorizationValidBefore PaymentReason = "invalid_exact_evm_payload_authorization_valid_before"
	PaymentReasonInvalidAuthorizationValue       PaymentReason = "invalid_exact_evm_payload_authorization_value_mismatch"
	PaymentReasonInvalidSignature                PaymentReason = "invalid_exact_evm_payload_signature"
	PaymentReasonInvalidRecipient                PaymentReason = "invalid_exact_evm_payload_recipient_mismatch"
	PaymentReasonInvalidNetwork                  PaymentReason = "invalid_network"
	PaymentReasonInvalidPayload                  PaymentReason = "invalid_payload"
	PaymentReasonInvalidPaymentRequirements      PaymentReason = "invalid_payment_requirements"
	PaymentReasonInvalidScheme                   PaymentReason = "invalid_scheme"
	PaymentReasonUnsupportedScheme               PaymentReason = "unsupported_scheme"
	PaymentReasonInvalidX402Version              PaymentReason = "invalid_x402_version"
	PaymentReasonInvalidTransactionState         PaymentReason = "invalid_transaction_state"
	PaymentReasonSettlementPending               PaymentReason = "settlement_pending"
)

func (r PaymentReason) String() string {
	return string(r)
}
