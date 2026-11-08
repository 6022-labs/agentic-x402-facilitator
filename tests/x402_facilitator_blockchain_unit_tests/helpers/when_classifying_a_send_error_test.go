package helpers_test

import (
	"errors"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenClassifyingASendError(t *testing.T) {
	t.Parallel()

	t.Run("Given the node already holds the transaction", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify it as already known", func(t *testing.T) {
			t.Parallel()

			kind := helpers.ClassifySendError(errors.New("already known"))

			assert.Equal(t, helpers.SendErrorAlreadyKnown, kind)
		})
	})

	t.Run("Given the nonce is already used on-chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify it as a taken nonce", func(t *testing.T) {
			t.Parallel()

			kind := helpers.ClassifySendError(errors.New("nonce too low: next nonce 8, tx nonce 7"))

			assert.Equal(t, helpers.SendErrorNonceTaken, kind)
		})
	})

	t.Run("Given another pending transaction uses the nonce", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify it as a taken nonce", func(t *testing.T) {
			t.Parallel()

			kind := helpers.ClassifySendError(errors.New("replacement transaction underpriced"))

			assert.Equal(t, helpers.SendErrorNonceTaken, kind)
		})
	})

	t.Run("Given the relayer cannot pay the gas", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify it as rejected", func(t *testing.T) {
			t.Parallel()

			kind := helpers.ClassifySendError(errors.New("insufficient funds for gas * price + value"))

			assert.Equal(t, helpers.SendErrorRejected, kind)
		})
	})

	t.Run("Given the fee is below what the node accepts", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify it as rejected", func(t *testing.T) {
			t.Parallel()

			kind := helpers.ClassifySendError(errors.New("transaction underpriced: tip needed 25000000000"))

			assert.Equal(t, helpers.SendErrorRejected, kind)
		})
	})

	t.Run("Given the fee is above the node's cap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify it as rejected", func(t *testing.T) {
			t.Parallel()

			kind := helpers.ClassifySendError(errors.New("tx fee (1.20 ether) exceeds the configured cap (1.00 ether)"))

			assert.Equal(t, helpers.SendErrorRejected, kind)
		})
	})

	t.Run("Given the request timed out", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify it as unknown", func(t *testing.T) {
			t.Parallel()

			kind := helpers.ClassifySendError(errors.New("Post \"https://rpc.example\": context deadline exceeded"))

			assert.Equal(t, helpers.SendErrorUnknown, kind)
		})
	})
}
