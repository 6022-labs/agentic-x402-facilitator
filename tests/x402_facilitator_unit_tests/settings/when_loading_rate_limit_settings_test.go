package settings_test

import (
	"testing"

	rate_limit_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingRateLimitSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured requests per second", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"rate_limit.requests_per_second": 120,
		}, "."), nil)

		t.Run("Should load it", func(t *testing.T) {
			t.Parallel()

			settings := rate_limit_settings.NewRateLimitSettings(zap.NewNop(), k)

			assert.NotNil(t, settings)
			assert.Equal(t, 120, settings.RequestsPerSecond)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should fall back to 50", func(t *testing.T) {
			t.Parallel()

			settings := rate_limit_settings.NewRateLimitSettings(zap.NewNop(), k)

			assert.NotNil(t, settings)
			assert.Equal(t, 50, settings.RequestsPerSecond)
		})
	})

	t.Run("Given a non-positive requests per second", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"rate_limit.requests_per_second": 0,
		}, "."), nil)

		t.Run("Should fall back to 50", func(t *testing.T) {
			t.Parallel()

			settings := rate_limit_settings.NewRateLimitSettings(zap.NewNop(), k)

			assert.NotNil(t, settings)
			assert.Equal(t, 50, settings.RequestsPerSecond)
		})
	})
}
