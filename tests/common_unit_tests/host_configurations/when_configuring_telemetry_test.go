package host_configurations_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/common/host_configurations"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type collectorRequest struct {
	path          string
	authorization string
}

type WhenConfiguringTelemetryTestingSuite struct {
	collector  *httptest.Server
	requestsMu sync.Mutex
	requests   []collectorRequest
	container  *dig.Container
}

func WhenConfiguringTelemetryBeforeEach(configuration map[string]any) *WhenConfiguringTelemetryTestingSuite {
	suite := &WhenConfiguringTelemetryTestingSuite{}
	suite.collector = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.requestsMu.Lock()
		suite.requests = append(suite.requests, collectorRequest{
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
		})
		suite.requestsMu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))

	configuration["telemetry.endpoint"] = strings.TrimPrefix(suite.collector.URL, "http://")
	k := koanf.New(".")
	_ = k.Load(confmap.Provider(configuration, "."), nil)

	suite.container = dig.New()
	_ = suite.container.Provide(func() *zap.Logger { return zap.NewNop() })
	_ = suite.container.Provide(func() *koanf.Koanf { return k })

	return suite
}

func WhenConfiguringTelemetryAfterEach(suite *WhenConfiguringTelemetryTestingSuite) {
	suite.collector.Close()
}

func (s *WhenConfiguringTelemetryTestingSuite) requestsTo(path string) []collectorRequest {
	s.requestsMu.Lock()
	defer s.requestsMu.Unlock()

	var matching []collectorRequest
	for _, request := range s.requests {
		if request.path == path {
			matching = append(matching, request)
		}
	}
	return matching
}

// ConfigureTelemetry sets the global OTel providers, so these tests never run in parallel.
func TestWhenConfiguringTelemetry(t *testing.T) {
	t.Run("Given the http protocol with an auth token", func(t *testing.T) {
		configuration := func() map[string]any {
			return map[string]any{
				"telemetry.enabled":     true,
				"telemetry.protocol":    "http",
				"telemetry.insecure":    true,
				"telemetry.instance_id": "node-1",
				"telemetry.environment": "test",
				"telemetry.auth_token":  "collector-token",
			}
		}

		t.Run("Should export spans to the traces path with the authorization header", func(t *testing.T) {
			suite := WhenConfiguringTelemetryBeforeEach(configuration())
			defer WhenConfiguringTelemetryAfterEach(suite)

			ctx := context.Background()
			shutdown := host_configurations.ConfigureTelemetry(suite.container, "test-service")
			defer func() { _ = shutdown(ctx) }()

			_, span := otel.Tracer("test").Start(ctx, "test-span")
			span.End()

			tracerProvider, ok := otel.GetTracerProvider().(interface{ ForceFlush(context.Context) error })
			require.True(t, ok)
			require.NoError(t, tracerProvider.ForceFlush(ctx))

			requests := suite.requestsTo("/v1/traces")
			require.NotEmpty(t, requests)
			assert.Equal(t, "Bearer collector-token", requests[0].authorization)
		})
	})
}
