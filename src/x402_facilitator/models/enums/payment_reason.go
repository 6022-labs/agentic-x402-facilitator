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
	PaymentReasonInvalidPermit2Spender           PaymentReason = "invalid_permit2_spender"
	PaymentReasonInvalidPermit2RecipientMismatch PaymentReason = "invalid_permit2_recipient_mismatch"
	PaymentReasonPermit2AmountMismatch           PaymentReason = "permit2_amount_mismatch"
	PaymentReasonPermit2TokenMismatch            PaymentReason = "permit2_token_mismatch"
	PaymentReasonInvalidPermit2Signature         PaymentReason = "invalid_permit2_signature"
	PaymentReasonPermit2AllowanceRequired        PaymentReason = "permit2_allowance_required"
	PaymentReasonPermit2ProxyNotDeployed         PaymentReason = "permit2_proxy_not_deployed"
)

func (r PaymentReason) String() string {
	return string(r)
}
