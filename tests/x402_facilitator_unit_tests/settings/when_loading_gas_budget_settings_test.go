package settings_test

import (
	"testing"

	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWhenLoadingGasBudgetSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain without a gas budget", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url": "https://rpc.example.com",
		}, "."), nil))

		t.Run("Should set no budget", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewGasBudgetSettings(zap.NewNop(), k)

			_, bounded := result.Get(137)
			assert.False(t, bounded)
		})
	})

	t.Run("Given a zero gas budget", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url":     "https://rpc.example.com",
			"chains.137.daily_gas_budget": 0,
		}, "."), nil))

		t.Run("Should set no budget", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewGasBudgetSettings(zap.NewNop(), k)

			_, bounded := result.Get(137)
			assert.False(t, bounded)
		})
	})

	t.Run("Given a whole gas budget", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url":     "https://rpc.example.com",
			"chains.137.daily_gas_budget": 5,
		}, "."), nil))

		t.Run("Should convert it to gwei", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewGasBudgetSettings(zap.NewNop(), k)

			budget, bounded := result.Get(137)
			assert.True(t, bounded)
			assert.Equal(t, int64(5_000_000_000), budget)
		})
	})

	t.Run("Given a fractional gas budget as a string", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url":     "https://rpc.example.com",
			"chains.137.daily_gas_budget": "0.25",
		}, "."), nil))

		t.Run("Should convert it to gwei", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewGasBudgetSettings(zap.NewNop(), k)

			budget, bounded := result.Get(137)
			assert.True(t, bounded)
			assert.Equal(t, int64(250_000_000), budget)
		})
	})

	t.Run("Given a negative gas budget", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url":     "https://rpc.example.com",
			"chains.137.daily_gas_budget": -1,
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() { x402_facilitator_settings.NewGasBudgetSettings(logger, k) })
		})
	})

	t.Run("Given a gas budget below one gwei", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url":     "https://rpc.example.com",
			"chains.137.daily_gas_budget": 0.0000000001,
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() { x402_facilitator_settings.NewGasBudgetSettings(logger, k) })
		})
	})

	t.Run("Given a gas budget above the largest countable one", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url":     "https://rpc.example.com",
			"chains.137.daily_gas_budget": 1e12,
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() { x402_facilitator_settings.NewGasBudgetSettings(logger, k) })
		})
	})

	t.Run("Given a gas budget that is not a number", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url":     "https://rpc.example.com",
			"chains.137.daily_gas_budget": "NaN",
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() { x402_facilitator_settings.NewGasBudgetSettings(logger, k) })
		})
	})
}
