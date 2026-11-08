package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingTheNetwork(t *testing.T) {
	t.Parallel()

	t.Run("Given the requirements are missing", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements = nil

		t.Run("Should return an empty network", func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, request.Network())
		})
	})

	t.Run("Given requirements naming a network", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Network = "polygon"

		t.Run("Should return it as sent", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "polygon", request.Network())
		})
	})
}
