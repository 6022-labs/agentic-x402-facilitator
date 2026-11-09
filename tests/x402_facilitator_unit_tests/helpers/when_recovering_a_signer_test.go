package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenRecoveringASigner(t *testing.T) {
	t.Parallel()

	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	digest := crypto.Keccak256([]byte("x402"))
	signature, err := crypto.Sign(digest, key)
	require.NoError(t, err)

	t.Run("Given a signature with a v of 0 or 1", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover its signer", func(t *testing.T) {
			t.Parallel()

			signer, err := helpers.RecoverSigner(digest, signature)

			require.NoError(t, err)
			assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey), signer)
		})
	})

	t.Run("Given a signature with a v of 27 or 28", func(t *testing.T) {
		t.Parallel()

		shifted := append([]byte(nil), signature...)
		shifted[64] += 27

		t.Run("Should recover its signer", func(t *testing.T) {
			t.Parallel()

			signer, err := helpers.RecoverSigner(digest, shifted)

			require.NoError(t, err)
			assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey), signer)
		})
	})
}
