package interfaces

import (
	"context"
	"time"
)

// ApiRequestMetricsInterface records http.server.* metrics for an inbound HTTP server.
type ApiRequestMetricsInterface interface {
	IncActiveRequests(ctx context.Context, method string, scheme string)
	DecActiveRequests(ctx context.Context, method string, scheme string)
	RecordRequest(ctx context.Context, method, route, scheme string, statusCode int, duration time.Duration)
}
