package enums

// PaymentScheme is an x402 payment scheme; only exact is settled.
type PaymentScheme string

const PaymentSchemeExact PaymentScheme = "exact"

func (s PaymentScheme) String() string {
	return string(s)
}
