package settings_test

import (
	"testing"
	"time"

	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWhenLoadingAccessSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no access settings", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should serve agents only", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewAccessSettings(zap.NewNop(), k)

			assert.True(t, result.AgentsOnly)
		})

		t.Run("Should cache lookups for five minutes", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewAccessSettings(zap.NewNop(), k)

			assert.Equal(t, 5*time.Minute, result.CacheTtl)
		})
	})

	t.Run("Given agents_only is the string false", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"access.agents_only": "false",
		}, "."), nil))

		t.Run("Should open access to every payee", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewAccessSettings(zap.NewNop(), k)

			assert.False(t, result.AgentsOnly)
		})
	})

	t.Run("Given a non-positive cache ttl", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"access.cache_ttl_seconds": 0,
		}, "."), nil))

		t.Run("Should default to five minutes", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewAccessSettings(zap.NewNop(), k)

			assert.Equal(t, 5*time.Minute, result.CacheTtl)
		})
	})
}
