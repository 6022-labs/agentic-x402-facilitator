package settings_test

import (
	"encoding/base64"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWhenLoadingTelemetrySettings(t *testing.T) {
	t.Parallel()

	t.Run("Given telemetry enabled with an endpoint, instance id and environment", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":     true,
			"telemetry.endpoint":    " collector:4317 ",
			"telemetry.insecure":    true,
			"telemetry.instance_id": " node-1 ",
			"telemetry.environment": " staging ",
		}, "."), nil)

		t.Run("Should load and trim the values", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.True(t, result.Enabled)
			assert.Equal(t, "collector:4317", result.Endpoint)
			assert.True(t, result.Insecure)
			assert.Equal(t, "node-1", result.InstanceId)
			assert.Equal(t, "staging", result.Environment)
		})

		t.Run("Should not authenticate", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.Empty(t, result.AuthScheme)

			header, ok := result.AuthorizationHeader()

			assert.False(t, ok)
			assert.Empty(t, header)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should stay disabled", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.False(t, result.Enabled)
			assert.Empty(t, result.Endpoint)
			assert.Empty(t, result.InstanceId)
			assert.Empty(t, result.Environment)
		})
	})

	t.Run("Given only an auth token", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":     true,
			"telemetry.endpoint":    "collector:4317",
			"telemetry.instance_id": "node-1",
			"telemetry.environment": "staging",
			"telemetry.auth_token":  " a-token ",
		}, "."), nil)

		t.Run("Should resolve bearer and send a bearer header", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.Equal(t, settings.TelemetryAuthSchemeBearer, result.AuthScheme)

			header, ok := result.AuthorizationHeader()

			assert.True(t, ok)
			assert.Equal(t, "Bearer a-token", header)
		})
	})

	t.Run("Given a username and a password", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":       true,
			"telemetry.endpoint":      "collector:4317",
			"telemetry.instance_id":   "node-1",
			"telemetry.environment":   "staging",
			"telemetry.auth_username": " ingest-user ",
			"telemetry.auth_password": " a-password ",
		}, "."), nil)

		t.Run("Should resolve basic and build the header from the credential, not the instance id", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.Equal(t, settings.TelemetryAuthSchemeBasic, result.AuthScheme)
			assert.Equal(t, "ingest-user", result.AuthUsername)

			header, ok := result.AuthorizationHeader()

			assert.True(t, ok)
			assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("ingest-user:a-password")), header)
		})
	})

	t.Run("Given auth settings while telemetry is disabled", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":    false,
			"telemetry.auth_token": "a-token",
		}, "."), nil)

		t.Run("Should not resolve a scheme", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.Empty(t, result.AuthScheme)

			_, ok := result.AuthorizationHeader()

			assert.False(t, ok)
		})
	})

	t.Run("Given no protocol", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should default to grpc", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.Equal(t, settings.TelemetryProtocolGrpc, result.Protocol)
		})
	})

	t.Run("Given the http protocol in mixed case", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":     true,
			"telemetry.protocol":    " HTTP ",
			"telemetry.endpoint":    "collector:4318",
			"telemetry.instance_id": "node-1",
			"telemetry.environment": "staging",
		}, "."), nil)

		t.Run("Should resolve the http protocol", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.Equal(t, settings.TelemetryProtocolHttp, result.Protocol)
			assert.Equal(t, "collector:4318", result.Endpoint)
		})
	})

	t.Run("Given an unknown protocol", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.protocol": "udp",
		}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() { settings.NewTelemetrySettings(logger, k) })
		})
	})

	t.Run("Given an endpoint with a scheme", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":     true,
			"telemetry.endpoint":    "http://collector:4318",
			"telemetry.instance_id": "node-1",
			"telemetry.environment": "staging",
		}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() { settings.NewTelemetrySettings(logger, k) })
		})
	})

	t.Run("Given an endpoint with a path", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":     true,
			"telemetry.endpoint":    "collector:4318/v1/traces",
			"telemetry.instance_id": "node-1",
			"telemetry.environment": "staging",
		}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() { settings.NewTelemetrySettings(logger, k) })
		})
	})
}
