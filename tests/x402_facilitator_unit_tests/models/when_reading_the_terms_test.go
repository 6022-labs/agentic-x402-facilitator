package models_test

import (
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingTheTerms(t *testing.T) {
	t.Parallel()

	t.Run("Given the same terms under another signature", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewTransferAuthorizationBuilder()
		authorization := builder.Build()
		other := builder.WithSignature([]byte{2}).Build()

		t.Run("Should read the same terms", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, authorization.Terms(), other.Terms())
		})
	})

	t.Run("Given another payee", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewTransferAuthorizationBuilder()
		authorization := builder.Build()
		other := builder.WithTo(common_tests.RandomEthAddress()).Build()

		t.Run("Should read other terms", func(t *testing.T) {
			t.Parallel()

			assert.NotEqual(t, authorization.Terms(), other.Terms())
		})
	})

	t.Run("Given another value", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewTransferAuthorizationBuilder()
		authorization := builder.Build()
		other := builder.WithValue(big.NewInt(1001)).Build()

		t.Run("Should read other terms", func(t *testing.T) {
			t.Parallel()

			assert.NotEqual(t, authorization.Terms(), other.Terms())
		})
	})

	t.Run("Given another validAfter", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewTransferAuthorizationBuilder()
		authorization := builder.Build()
		other := builder.WithValidAfter(big.NewInt(11)).Build()

		t.Run("Should read other terms", func(t *testing.T) {
			t.Parallel()

			assert.NotEqual(t, authorization.Terms(), other.Terms())
		})
	})

	t.Run("Given another validBefore", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewTransferAuthorizationBuilder()
		authorization := builder.Build()
		other := builder.WithValidBefore(big.NewInt(21)).Build()

		t.Run("Should read other terms", func(t *testing.T) {
			t.Parallel()

			assert.NotEqual(t, authorization.Terms(), other.Terms())
		})
	})
}
