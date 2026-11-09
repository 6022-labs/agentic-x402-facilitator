package helpers_test

import (
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingAUint256(t *testing.T) {
	t.Parallel()

	t.Run("Given a negative integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			_, ok := helpers.ParseUint256("-1")

			assert.False(t, ok)
		})
	})

	t.Run("Given an integer past 256 bits", func(t *testing.T) {
		t.Parallel()

		tooLarge := new(big.Int).Lsh(big.NewInt(1), 256).String()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			_, ok := helpers.ParseUint256(tooLarge)

			assert.False(t, ok)
		})
	})

	t.Run("Given a decimal integer with surrounding spaces", func(t *testing.T) {
		t.Parallel()

		t.Run("Should parse it", func(t *testing.T) {
			t.Parallel()

			value, ok := helpers.ParseUint256(" 42 ")

			assert.True(t, ok)
			assert.Equal(t, big.NewInt(42), value)
		})
	})
}
