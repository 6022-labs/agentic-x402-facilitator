package middlewares

import (
	"time"

	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/common/metrics/interfaces"
	"github.com/gofiber/fiber/v3"
)

// ApiRequestMetricsMiddleware records http.server.* metrics around every request; must run before handlers.
type ApiRequestMetricsMiddleware struct {
	metrics metrics_interfaces.ApiRequestMetricsInterface
}

func NewApiRequestMetricsMiddleware(
	metrics metrics_interfaces.ApiRequestMetricsInterface,
) *ApiRequestMetricsMiddleware {
	return &ApiRequestMetricsMiddleware{
		metrics: metrics,
	}
}

func (m *ApiRequestMetricsMiddleware) Handle(ctx fiber.Ctx) error {
	start := time.Now()
	method := ctx.Method()

	// Scheme() reads forwarded headers only from trusted proxies and yields only http or https.
	scheme := ctx.Scheme()

	userCtx := ctx.Context()
	m.metrics.IncActiveRequests(userCtx, method, scheme)
	err := ctx.Next()
	m.metrics.DecActiveRequests(userCtx, method, scheme)
	if err != nil {
		_ = ctx.App().ErrorHandler(ctx, err)
		err = nil
	}

	// Prefer the route template (low cardinality) over the raw path.
	route := ctx.Path()
	if ctx.Route() != nil && ctx.Route().Path != "" {
		route = ctx.Route().Path
	}

	m.metrics.RecordRequest(userCtx, method, route, scheme, ctx.Response().StatusCode(), time.Since(start))

	return err
}
