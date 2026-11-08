package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingTheAcceptedNetwork(t *testing.T) {
	t.Parallel()

	t.Run("Given a v2 payload naming its accepted requirements", func(t *testing.T) {
		t.Parallel()

		payload := &requests.PaymentPayload{
			Accepted: &requests.PaymentRequirements{Network: "eip155:80002"},
			Network:  "ignored",
		}

		t.Run("Should read it from accepted", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "eip155:80002", payload.AcceptedNetwork())
		})
	})

	t.Run("Given a legacy payload with a top-level network", func(t *testing.T) {
		t.Parallel()

		payload := &requests.PaymentPayload{Network: "eip155:137"}

		t.Run("Should read the top-level field", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "eip155:137", payload.AcceptedNetwork())
		})
	})
}
