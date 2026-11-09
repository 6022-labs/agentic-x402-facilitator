package settings_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WhenLoadingTokensSettingsTestingSuite struct{}

func (suite *WhenLoadingTokensSettingsTestingSuite) panicLogger() *zap.Logger {
	return zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))
}

func TestWhenLoadingTokensSettings(t *testing.T) {
	t.Parallel()

	suite := &WhenLoadingTokensSettingsTestingSuite{}
	amoyUsdc := common_tests.RandomEthAddress().Hex()

	t.Run("Given no token is configured", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given a chain listed without tokens", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002": map[string]any{},
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given a token keyed by something other than an address", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002.usdc.eip712_name":    "USDC",
			"tokens.80002.usdc.eip712_version": "2",
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given a token entry set to null", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002": map[string]any{amoyUsdc: nil},
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given a token set in JSON and in env with different casing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + amoyUsdc + ".eip712_name":                     "USDC",
			"tokens.80002." + amoyUsdc + ".eip712_version":                  "2",
			"tokens.80002." + strings.ToLower(amoyUsdc) + ".eip712_name":    "USD Coin",
			"tokens.80002." + strings.ToLower(amoyUsdc) + ".eip712_version": "2",
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given a token with an unknown asset transfer method", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + amoyUsdc + ".asset_transfer_methods": "permit2,erc7710",
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given an eip3009 token without its EIP-712 name", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + amoyUsdc + ".asset_transfer_methods": "eip3009",
			"tokens.80002." + amoyUsdc + ".eip712_version":         "2",
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given an eip3009 token without its EIP-712 version", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + amoyUsdc + ".asset_transfer_methods": "eip3009",
			"tokens.80002." + amoyUsdc + ".eip712_name":            "USDC",
		}, "."), nil))

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				x402_facilitator_settings.NewTokensSettings(suite.panicLogger(), k)
			})
		})
	})

	t.Run("Given a token without asset transfer methods", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + amoyUsdc + ".eip712_name":    "USDC",
			"tokens.80002." + amoyUsdc + ".eip712_version": "2",
		}, "."), nil))

		t.Run("Should settle it with eip3009", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewTokensSettings(zap.NewNop(), k)

			token := result.Get(80002, common.HexToAddress(amoyUsdc))
			require.NotNil(t, token)
			assert.Equal(t, []enums.AssetTransferMethod{enums.AssetTransferMethodEip3009}, token.AssetTransferMethods)
		})
	})

	t.Run("Given a permit2 token without an EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + amoyUsdc + ".asset_transfer_methods": "permit2",
		}, "."), nil))

		t.Run("Should load it", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewTokensSettings(zap.NewNop(), k)

			token := result.Get(80002, common.HexToAddress(amoyUsdc))
			require.NotNil(t, token)
			assert.Equal(t, []enums.AssetTransferMethod{enums.AssetTransferMethodPermit2}, token.AssetTransferMethods)
		})
	})

	t.Run("Given a token keyed by chain id and address", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + amoyUsdc + ".asset_transfer_methods": "eip3009, permit2",
			"tokens.80002." + amoyUsdc + ".eip712_name":            " USDC ",
			"tokens.80002." + amoyUsdc + ".eip712_version":         "2",
		}, "."), nil))

		t.Run("Should expose its methods", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewTokensSettings(zap.NewNop(), k)

			token := result.Get(80002, common.HexToAddress(amoyUsdc))
			require.NotNil(t, token)
			assert.Equal(
				t,
				[]enums.AssetTransferMethod{enums.AssetTransferMethodEip3009, enums.AssetTransferMethodPermit2},
				token.AssetTransferMethods,
			)
		})

		t.Run("Should expose its trimmed EIP-712 domain", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewTokensSettings(zap.NewNop(), k)

			token := result.Get(80002, common.HexToAddress(amoyUsdc))
			require.NotNil(t, token)
			assert.Equal(t, "USDC", token.Eip712Name)
			assert.Equal(t, "2", token.Eip712Version)
		})
	})

	t.Run("Given a token keyed by a lowercased address", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"tokens.80002." + strings.ToLower(amoyUsdc) + ".asset_transfer_methods": "permit2",
		}, "."), nil))

		t.Run("Should find it by its checksummed address", func(t *testing.T) {
			t.Parallel()

			result := x402_facilitator_settings.NewTokensSettings(zap.NewNop(), k)

			assert.NotNil(t, result.Get(80002, common.HexToAddress(amoyUsdc)))
		})
	})
}
