package settings_test

import (
	"testing"

	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingAgentCollectionsManagersSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given the built-in addresses", func(t *testing.T) {
		t.Parallel()

		t.Run("Should know Polygon and Polygon Amoy", func(t *testing.T) {
			t.Parallel()

			result := blockchain_settings.NewAgentCollectionsManagersSettings(zap.NewNop())

			polygon, found := result.Get(137)
			assert.True(t, found)
			assert.Equal(t, common.HexToAddress("0xb0dc8a83c700A9BBcc53cA1a2C6993a63129d2F6"), polygon)
			amoy, found := result.Get(80002)
			assert.True(t, found)
			assert.Equal(t, common.HexToAddress("0x8095A0D4bE42A61db9e609aA693f516044Bda990"), amoy)
		})

		t.Run("Should know no other chain", func(t *testing.T) {
			t.Parallel()

			result := blockchain_settings.NewAgentCollectionsManagersSettings(zap.NewNop())

			_, found := result.Get(1)
			assert.False(t, found)
		})
	})
}
