package settings_test

import (
	"testing"

	host_settings "github.com/6022-labs/agentic-x402-facilitator/src/common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingHostSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a complete configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"host.api_port":         8080,
			"host.body_limit_bytes": 1024,
			"host.listen_address":   "127.0.0.1",
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := host_settings.NewHostSettings(zap.NewNop(), k, "host", 5000)

			assert.Equal(t, 8080, result.ApiPort)
			assert.Equal(t, 1024, result.BodyLimitBytes)
			assert.Equal(t, "127.0.0.1", result.ListenAddress)
		})
	})

	t.Run("Given api_port is missing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{}, "."), nil)

		t.Run("Should default to 5000", func(t *testing.T) {
			t.Parallel()

			result := host_settings.NewHostSettings(zap.NewNop(), k, "host", 5000)

			assert.Equal(t, 5000, result.ApiPort)
		})
	})

	t.Run("Given body_limit_bytes is missing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{}, "."), nil)

		t.Run("Should default to 4MB and bind every interface", func(t *testing.T) {
			t.Parallel()

			result := host_settings.NewHostSettings(zap.NewNop(), k, "host", 5000)

			assert.Equal(t, 4*1024*1024, result.BodyLimitBytes)
			assert.Equal(t, "", result.ListenAddress)
		})
	})

	t.Run("Given a non-positive api port", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"host.api_port": 0,
		}, "."), nil)

		t.Run("Should clamp it back to the default", func(t *testing.T) {
			t.Parallel()

			result := host_settings.NewHostSettings(zap.NewNop(), k, "host", 5000)

			assert.Equal(t, 5000, result.ApiPort)
		})
	})

	t.Run("Given a non-positive body limit", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"host.body_limit_bytes": 0,
		}, "."), nil)

		t.Run("Should clamp it back to 4MB", func(t *testing.T) {
			t.Parallel()

			result := host_settings.NewHostSettings(zap.NewNop(), k, "host", 5000)

			assert.Equal(t, 4*1024*1024, result.BodyLimitBytes)
		})
	})
}
