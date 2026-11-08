package metrics

import (
	"context"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type ApiRequestMetrics struct {
	durationHistogram metric.Float64Histogram
	activeRequests    metric.Int64UpDownCounter
}

// NewApiRequestMetrics builds the http.server.* instruments under the given meter scope.
// The convention defines no request/error counters: the duration histogram's count is the request
// count, and failures are the error.type dimension on it.
func NewApiRequestMetrics(meterName string) *ApiRequestMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	durationHistogram, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencySecondsBuckets...),
		metric.WithDescription("Duration of HTTP server requests"),
	)
	if err != nil {
		durationHistogram, _ = noopMeter.Float64Histogram("http.server.request.duration")
	}

	activeRequests, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithUnit("{request}"),
		metric.WithDescription("Number of active HTTP server requests"),
	)
	if err != nil {
		activeRequests, _ = noopMeter.Int64UpDownCounter("http.server.active_requests")
	}

	return &ApiRequestMetrics{
		durationHistogram: durationHistogram,
		activeRequests:    activeRequests,
	}
}

func (m *ApiRequestMetrics) IncActiveRequests(ctx context.Context, method string, scheme string) {
	m.activeRequests.Add(ctx, 1, m.activeAttributes(method, scheme))
}

func (m *ApiRequestMetrics) DecActiveRequests(ctx context.Context, method string, scheme string) {
	m.activeRequests.Add(ctx, -1, m.activeAttributes(method, scheme))
}

func (m *ApiRequestMetrics) RecordRequest(
	ctx context.Context,
	method, route, scheme string,
	statusCode int,
	duration time.Duration,
) {
	attrs := []attribute.KeyValue{
		attribute.String("http.request.method", method),
		attribute.String("http.route", route),
		attribute.String("url.scheme", scheme),
		attribute.Int("http.response.status_code", statusCode),
	}
	// error.type is conditionally required, and only on a failed request.
	if statusCode >= 500 {
		attrs = append(attrs, attribute.String("error.type", strconv.Itoa(statusCode)))
	}
	m.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *ApiRequestMetrics) activeAttributes(method string, scheme string) metric.MeasurementOption {
	return metric.WithAttributes(
		attribute.String("http.request.method", method),
		attribute.String("url.scheme", scheme),
	)
}
