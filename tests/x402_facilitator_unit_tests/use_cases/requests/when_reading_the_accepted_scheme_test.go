package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingTheAcceptedScheme(t *testing.T) {
	t.Parallel()

	t.Run("Given a v2 payload naming its accepted requirements", func(t *testing.T) {
		t.Parallel()

		payload := &requests.PaymentPayload{
			Accepted: &requests.PaymentRequirements{Scheme: "exact"},
			Scheme:   "ignored",
		}

		t.Run("Should read it from accepted", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "exact", payload.AcceptedScheme())
		})
	})

	t.Run("Given a legacy payload with a top-level scheme", func(t *testing.T) {
		t.Parallel()

		payload := &requests.PaymentPayload{Scheme: "exact"}

		t.Run("Should read the top-level field", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "exact", payload.AcceptedScheme())
		})
	})
}
