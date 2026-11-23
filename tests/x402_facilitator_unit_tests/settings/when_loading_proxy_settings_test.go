package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWhenLoadingProxySettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trust no proxy", func(t *testing.T) {
			t.Parallel()

			result := settings.NewProxySettings(zap.NewNop(), koanf.New("."))

			assert.Empty(t, result.Header)
			assert.Empty(t, result.TrustedProxies)
		})
	})

	t.Run("Given a header and trusted proxies", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"proxy.header":          " X-Forwarded-For ",
			"proxy.trusted_proxies": []string{"10.0.0.0/8", " 192.168.1.10"},
		}, "."), nil)

		t.Run("Should load them trimmed", func(t *testing.T) {
			t.Parallel()

			result := settings.NewProxySettings(zap.NewNop(), k)

			assert.Equal(t, "X-Forwarded-For", result.Header)
			assert.Equal(t, []string{"10.0.0.0/8", "192.168.1.10"}, result.TrustedProxies)
		})
	})

	t.Run("Given trusted proxies as one comma-separated env value", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"proxy.header":          "X-Real-IP",
			"proxy.trusted_proxies": "10.0.0.0/8, fd00::/8,",
		}, "."), nil)

		t.Run("Should split them", func(t *testing.T) {
			t.Parallel()

			result := settings.NewProxySettings(zap.NewNop(), k)

			assert.Equal(t, []string{"10.0.0.0/8", "fd00::/8"}, result.TrustedProxies)
		})
	})

	t.Run("Given a header without trusted proxies", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"proxy.header": "X-Forwarded-For"}, "."), nil)

		t.Run("Should drop the header", func(t *testing.T) {
			t.Parallel()

			result := settings.NewProxySettings(zap.NewNop(), k)

			assert.Empty(t, result.Header)
		})
	})

	t.Run("Given an invalid trusted proxy", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"proxy.header":          "X-Forwarded-For",
			"proxy.trusted_proxies": []string{"proxy.internal"},
		}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() {
				settings.NewProxySettings(logger, k)
			})
		})
	})
}
