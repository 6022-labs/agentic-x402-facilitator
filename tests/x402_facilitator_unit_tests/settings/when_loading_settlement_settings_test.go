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

func TestWhenLoadingSettlementSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no settlement settings", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should lock a settlement for ten minutes", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewSettlementSettings(zap.NewNop(), k)

			assert.Equal(t, 10*time.Minute, result.LockTtl)
		})
	})

	t.Run("Given a lock ttl as a string", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"settlement.lock_ttl_seconds": "120",
		}, "."), nil))

		t.Run("Should use it", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewSettlementSettings(zap.NewNop(), k)

			assert.Equal(t, 2*time.Minute, result.LockTtl)
		})
	})

	t.Run("Given a non-positive lock ttl", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"settlement.lock_ttl_seconds": 0,
		}, "."), nil))

		t.Run("Should default to ten minutes", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewSettlementSettings(zap.NewNop(), k)

			assert.Equal(t, 10*time.Minute, result.LockTtl)
		})
	})
}
