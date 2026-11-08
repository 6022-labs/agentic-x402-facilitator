package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingASettlementKey(t *testing.T) {
	t.Parallel()

	t.Run("Given two payments of the same authorization", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		first := builder.BuildPayment()
		second := builder.BuildPayment()

		t.Run("Should give them the same key", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, first.SettlementKey(), second.SettlementKey())
		})
	})

	t.Run("Given two payments with different nonces", func(t *testing.T) {
		t.Parallel()

		first := model_builders.NewPaymentRequestBuilder().BuildPayment()
		second := model_builders.NewPaymentRequestBuilder().BuildPayment()

		t.Run("Should give them different keys", func(t *testing.T) {
			t.Parallel()

			assert.NotEqual(t, first.SettlementKey(), second.SettlementKey())
		})
	})
}
