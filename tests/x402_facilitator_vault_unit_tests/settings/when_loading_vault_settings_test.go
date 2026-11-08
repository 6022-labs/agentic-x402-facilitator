package settings_test

import (
	"testing"

	vault_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WhenLoadingVaultSettingsTestingSuite struct{}

func (suite *WhenLoadingVaultSettingsTestingSuite) panicLogger() *zap.Logger {
	return zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))
}

func TestWhenLoadingVaultSettings(t *testing.T) {
	t.Parallel()

	suite := &WhenLoadingVaultSettingsTestingSuite{}

	t.Run("Given the address is missing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"vault.token": "token-123"}, "."), nil)

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				vault_settings.NewVaultSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given the token is missing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"vault.address": "http://localhost:8200"}, "."), nil)

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				vault_settings.NewVaultSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given kv_v1_mount_path is missing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"vault.address": "http://localhost:8200",
			"vault.token":   "token-123",
		}, "."), nil)

		t.Run("Should default kv_v1_mount_path to 'secret'", func(t *testing.T) {
			t.Parallel()

			result := vault_settings.NewVaultSettings(zap.NewNop(), k)

			assert.Equal(t, "secret", result.KvV1MountPath)
		})
	})

	t.Run("Given padded values", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"vault.address":          "  http://vault:8200  ",
			"vault.token":            "  token  ",
			"vault.kv_v1_mount_path": "  custom  ",
		}, "."), nil)

		t.Run("Should trim every value", func(t *testing.T) {
			t.Parallel()

			result := vault_settings.NewVaultSettings(zap.NewNop(), k)

			assert.Equal(t, "http://vault:8200", result.Address)
			assert.Equal(t, "token", result.Token)
			assert.Equal(t, "custom", result.KvV1MountPath)
		})
	})

	t.Run("Given a complete configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"vault.address":          "http://localhost:8200",
			"vault.token":            "token-123",
			"vault.kv_v1_mount_path": "secret",
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := vault_settings.NewVaultSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "http://localhost:8200", result.Address)
			assert.Equal(t, "token-123", result.Token)
			assert.Equal(t, "secret", result.KvV1MountPath)
		})
	})
}
