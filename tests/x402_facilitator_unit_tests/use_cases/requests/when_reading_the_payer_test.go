package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingThePayer(t *testing.T) {
	t.Parallel()

	t.Run("Given the payload has no authorization", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization = nil

		t.Run("Should return an empty payer", func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, request.Payer())
		})
	})

	t.Run("Given the authorization from is not an address", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization.From = "payer"

		t.Run("Should return an empty payer", func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, request.Payer())
		})
	})

	t.Run("Given a lowercased authorization from", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.Build()
		request.PaymentPayload.Payload.Authorization.From = strings.ToLower(builder.Payer().Hex())

		t.Run("Should return it checksummed", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, builder.Payer().Hex(), request.Payer())
		})
	})
}
