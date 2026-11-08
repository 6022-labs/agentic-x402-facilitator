package settings_test

import (
	"testing"
	"time"

	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WhenLoadingChainsSettingsTestingSuite struct {
	tokensSettings     *x402_facilitator_settings.TokensSettings
	settlementSettings *x402_facilitator_settings.SettlementSettings
}

func WhenLoadingChainsSettingsBeforeEach() *WhenLoadingChainsSettingsTestingSuite {
	return &WhenLoadingChainsSettingsTestingSuite{
		settlementSettings: &x402_facilitator_settings.SettlementSettings{LockTtl: 10 * time.Minute},
		tokensSettings: &x402_facilitator_settings.TokensSettings{
			80002: {common_tests.RandomEthAddress(): {Eip712Name: "USDC", Eip712Version: "2"}},
		},
	}
}

func (suite *WhenLoadingChainsSettingsTestingSuite) panicLogger() *zap.Logger {
	return zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))
}

func TestWhenLoadingChainsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain entry set to null", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains": map[string]any{"80002": nil},
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			assert.Panics(t, func() {
				blockchain_settings.NewChainsSettings(
					suite.panicLogger(),
					k,
					suite.tokensSettings,
					suite.settlementSettings,
				)
			})
		})
	})

	t.Run("Given a chain without an rpc url", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.80002.confirmations": 2,
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			assert.Panics(t, func() {
				blockchain_settings.NewChainsSettings(
					suite.panicLogger(),
					k,
					suite.tokensSettings,
					suite.settlementSettings,
				)
			})
		})
	})

	t.Run("Given a chain without its testnet flag", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.80002.rpc_http_url": "https://polygon-amoy.example",
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			assert.Panics(t, func() {
				blockchain_settings.NewChainsSettings(
					suite.panicLogger(),
					k,
					suite.tokensSettings,
					suite.settlementSettings,
				)
			})
		})
	})

	t.Run("Given a settlement could outlive its lock", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.80002.rpc_http_url":            "https://polygon-amoy.example",
			"chains.80002.receipt_timeout_seconds": 600,
			"chains.80002.testnet":                 true,
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			assert.Panics(t, func() {
				blockchain_settings.NewChainsSettings(
					suite.panicLogger(),
					k,
					suite.tokensSettings,
					suite.settlementSettings,
				)
			})
		})
	})

	t.Run("Given a token chain without an rpc url", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.137.rpc_http_url": "https://polygon.example",
			"chains.137.testnet":      false,
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			assert.Panics(t, func() {
				blockchain_settings.NewChainsSettings(
					suite.panicLogger(),
					k,
					suite.tokensSettings,
					suite.settlementSettings,
				)
			})
		})
	})

	t.Run("Given a chain with only its required settings", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.80002.rpc_http_url": "  https://polygon-amoy.example  ",
			"chains.80002.testnet":      "false",
		}, "."), nil))

		t.Run("Should trim the url", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			result := blockchain_settings.NewChainsSettings(
				zap.NewNop(),
				k,
				suite.tokensSettings,
				suite.settlementSettings,
			)

			assert.Equal(t, "https://polygon-amoy.example", result.Get(80002).RpcHttpUrl)
		})

		t.Run("Should apply the default timeouts and confirmations", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			result := blockchain_settings.NewChainsSettings(
				zap.NewNop(),
				k,
				suite.tokensSettings,
				suite.settlementSettings,
			)

			chain := result.Get(80002)
			require.NotNil(t, chain)
			assert.Equal(t, 10*time.Second, chain.RequestTimeout)
			assert.Equal(t, uint64(1), chain.Confirmations)
			assert.Equal(t, 120*time.Second, chain.ReceiptTimeout)
			assert.False(t, chain.Testnet)
		})
	})

	t.Run("Given a chain with explicit settlement settings", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"chains.80002.rpc_http_url":            "https://polygon-amoy.example",
			"chains.80002.request_timeout_seconds": "5",
			"chains.80002.confirmations":           "3",
			"chains.80002.receipt_timeout_seconds": "60",
			"chains.80002.testnet":                 "true",
		}, "."), nil))

		t.Run("Should keep them", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()

			result := blockchain_settings.NewChainsSettings(
				zap.NewNop(),
				k,
				suite.tokensSettings,
				suite.settlementSettings,
			)

			chain := result.Get(80002)
			require.NotNil(t, chain)
			assert.Equal(t, 5*time.Second, chain.RequestTimeout)
			assert.Equal(t, uint64(3), chain.Confirmations)
			assert.Equal(t, 60*time.Second, chain.ReceiptTimeout)
			assert.True(t, chain.Testnet)
		})
	})
}
