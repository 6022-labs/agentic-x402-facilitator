package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenSplittingASignature(t *testing.T) {
	t.Parallel()

	t.Run("Given a signature whose v is 0 or 1", func(t *testing.T) {
		t.Parallel()

		signature := make([]byte, 65)
		signature[64] = 1

		t.Run("Should shift v to 27 or 28", func(t *testing.T) {
			t.Parallel()

			v, _, _ := helpers.SplitSignature(signature)

			assert.Equal(t, uint8(28), v)
		})
	})

	t.Run("Given a signature whose v is 27 or 28", func(t *testing.T) {
		t.Parallel()

		signature := make([]byte, 65)
		for i := range signature {
			signature[i] = byte(i)
		}
		signature[64] = 27

		t.Run("Should keep v", func(t *testing.T) {
			t.Parallel()

			v, _, _ := helpers.SplitSignature(signature)

			assert.Equal(t, uint8(27), v)
		})

		t.Run("Should split r and s", func(t *testing.T) {
			t.Parallel()

			_, r, s := helpers.SplitSignature(signature)

			assert.Equal(t, signature[0:32], r[:])
			assert.Equal(t, signature[32:64], s[:])
		})
	})
}
