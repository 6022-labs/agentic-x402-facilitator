package settings_test

import (
	"testing"

	redis_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestWhenLoadingRedisSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a complete configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"redis.db":       3,
			"redis.host":     " localhost:6379 ",
			"redis.password": " secret ",
		}, "."), nil)

		t.Run("Should load every field and trim the host and password", func(t *testing.T) {
			t.Parallel()

			result := redis_settings.NewRedisSettings(zap.NewNop(), k)

			assert.Equal(t, int64(3), result.DB)
			assert.Equal(t, "localhost:6379", result.Host)
			assert.Equal(t, "secret", result.Password)
		})
	})

	t.Run("Given the password is missing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"redis.host": "localhost:6379",
		}, "."), nil)

		t.Run("Should warn but still load the settings", func(t *testing.T) {
			t.Parallel()

			core, logs := observer.New(zap.WarnLevel)
			result := redis_settings.NewRedisSettings(zap.New(core), k)

			assert.Equal(t, "localhost:6379", result.Host)
			assert.Equal(t, 1, logs.FilterMessage("redis password is not set").Len())
		})
	})

	t.Run("Given a negative db index", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"redis.db":       -1,
			"redis.host":     "localhost:6379",
			"redis.password": "secret",
		}, "."), nil)

		t.Run("Should default the db index to 0", func(t *testing.T) {
			t.Parallel()

			result := redis_settings.NewRedisSettings(zap.NewNop(), k)

			assert.Equal(t, int64(0), result.DB)
		})
	})
}
