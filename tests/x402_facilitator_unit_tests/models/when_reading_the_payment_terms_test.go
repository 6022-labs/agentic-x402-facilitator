package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingThePaymentTerms(t *testing.T) {
	t.Parallel()

	t.Run("Given the same authorization for another token", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder().WithPermit2()
		first := builder.BuildPayment()
		second := builder.WithAsset(common_tests.RandomEthAddress()).BuildPayment()

		t.Run("Should read other terms", func(t *testing.T) {
			t.Parallel()

			assert.NotEqual(t, first.Terms(), second.Terms())
		})
	})

	t.Run("Given the same authorization twice", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()

		t.Run("Should read the same terms", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, builder.BuildPayment().Terms(), builder.BuildPayment().Terms())
		})
	})
}
