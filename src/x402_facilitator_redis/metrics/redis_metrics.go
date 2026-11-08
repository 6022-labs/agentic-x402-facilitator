package metrics

import (
	"context"
	"time"

	common_metrics "github.com/6022-labs/agentic-x402-facilitator/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

const meterName = "agentic_x402_facilitator_redis"

// RedisMetrics reports redis under the database semconv — redis is a db.system.name, not a bespoke cache tier.
type RedisMetrics struct {
	durationHistogram metric.Float64Histogram
}

func NewRedisMetrics() *RedisMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	durationHistogram, err := meter.Float64Histogram(
		"db.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.DbLatencySecondsBuckets...),
		metric.WithDescription("Duration of redis commands"),
	)
	if err != nil {
		durationHistogram, _ = noopMeter.Float64Histogram("db.client.operation.duration")
	}

	return &RedisMetrics{
		durationHistogram: durationHistogram,
	}
}

func (m *RedisMetrics) RecordOperation(ctx context.Context, operation string, duration time.Duration, err error) {
	attrs := []attribute.KeyValue{
		semconv.DBSystemNameRedis,
		attribute.String("db.operation.name", operation),
	}
	// error.type is conditionally required, and only on a failed operation.
	if err != nil {
		attrs = append(attrs, attribute.String("error.type", common_metrics.ErrorType(err)))
	}

	m.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}
