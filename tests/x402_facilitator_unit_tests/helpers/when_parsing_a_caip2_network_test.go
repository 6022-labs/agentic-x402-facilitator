package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingACaip2Network(t *testing.T) {
	t.Parallel()

	t.Run("Given another namespace", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			_, err := helpers.ParseCaip2("solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp")

			assert.ErrorContains(t, err, "must be a CAIP-2 eip155 identifier")
		})
	})

	t.Run("Given chain id zero", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			_, err := helpers.ParseCaip2("eip155:0")

			assert.ErrorContains(t, err, "must be a CAIP-2 eip155 identifier")
		})
	})
	t.Run("Given an eip155 identifier", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its chain id", func(t *testing.T) {
			t.Parallel()

			chainId, err := helpers.ParseCaip2("eip155:80002")

			assert.NoError(t, err)
			assert.Equal(t, uint64(80002), chainId)
		})
	})
}
