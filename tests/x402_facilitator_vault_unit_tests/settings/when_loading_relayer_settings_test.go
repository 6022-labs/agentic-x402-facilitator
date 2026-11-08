package settings_test

import (
	"testing"

	vault_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWhenLoadingRelayerSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no vault key is set", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should default the vault key to relayer", func(t *testing.T) {
			t.Parallel()

			result := vault_settings.NewRelayerSettings(zap.NewNop(), k)

			assert.Equal(t, "relayer", result.VaultKey)
		})
	})

	t.Run("Given a blank vault key", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"relayer.vault_key": "   ",
		}, "."), nil))

		t.Run("Should default the vault key to relayer", func(t *testing.T) {
			t.Parallel()

			result := vault_settings.NewRelayerSettings(zap.NewNop(), k)

			assert.Equal(t, "relayer", result.VaultKey)
		})
	})

	t.Run("Given a vault key with surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"relayer.vault_key": "  x402/relayer  ",
		}, "."), nil))

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			result := vault_settings.NewRelayerSettings(zap.NewNop(), k)

			assert.Equal(t, "x402/relayer", result.VaultKey)
		})
	})
}
