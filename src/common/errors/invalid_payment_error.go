package errors

import "fmt"

// InvalidPaymentError is a payment the facilitator refuses; Reason is an x402 reason code, not an HTTP failure.
type InvalidPaymentError struct {
	Reason  string
	Message string
}

func NewInvalidPaymentError[R ~string](reason R, message string) *InvalidPaymentError {
	return &InvalidPaymentError{Reason: string(reason), Message: message}
}

func (e *InvalidPaymentError) Error() string {
	return fmt.Sprintf("%s: %s", e.Reason, e.Message)
}
