package metrics

import (
	"context"
	"time"

	common_metrics "github.com/6022-labs/agentic-x402-facilitator/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_x402_facilitator_vault"

type VaultMetrics struct {
	durationHistogram metric.Float64Histogram
}

func NewVaultMetrics() *VaultMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	durationHistogram, err := meter.Float64Histogram(
		"vault.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LatencySecondsBuckets...),
		metric.WithDescription("Duration of Vault secret operations"),
	)
	if err != nil {
		durationHistogram, _ = noopMeter.Float64Histogram("vault.operation.duration")
	}

	return &VaultMetrics{
		durationHistogram: durationHistogram,
	}
}

// RecordOperation records one Vault operation: count, duration, and (when failed) an error.
func (m *VaultMetrics) RecordOperation(
	ctx context.Context,
	operation, mount string,
	duration time.Duration,
	err error,
) {
	attrs := []attribute.KeyValue{
		attribute.String("vault.operation.name", operation),
		attribute.String("vault.mount", mount),
	}
	// error.type is conditionally required, and only on a failed operation.
	if err != nil {
		attrs = append(attrs, attribute.String("error.type", common_metrics.ErrorType(err)))
	}

	m.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}
