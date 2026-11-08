package metrics

import (
	"context"
	"strconv"
	"time"

	common_metrics "github.com/6022-labs/agentic-x402-facilitator/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

const meterName = "agentic_x402_facilitator_blockchain"

type BlockchainRpcMetrics struct {
	durationHistogram metric.Float64Histogram
}

func NewBlockchainRpcMetrics() *BlockchainRpcMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	durationHistogram, err := meter.Float64Histogram(
		"rpc.client.duration",
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(common_metrics.RpcLatencyMillisecondsBuckets...),
		metric.WithDescription("Duration of blockchain JSON-RPC requests"),
	)
	if err != nil {
		durationHistogram, _ = noopMeter.Float64Histogram("rpc.client.duration")
	}

	return &BlockchainRpcMetrics{
		durationHistogram: durationHistogram,
	}
}

// RecordRequest records one JSON-RPC call: count, duration, and (when failed) an error.
func (m *BlockchainRpcMetrics) RecordRequest(
	ctx context.Context,
	chainId string,
	serverUrl string,
	duration time.Duration,
	statusCode int,
	err error,
) {
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("jsonrpc"),
		attribute.String("chain.id", chainId),
	}
	attrs = append(attrs, common_metrics.ServerAttributes(serverUrl)...)
	// error.type is conditionally required, and only on a failed call.
	switch {
	case err != nil:
		attrs = append(attrs, attribute.String("error.type", common_metrics.ErrorType(err)))
	case statusCode >= 400:
		attrs = append(attrs, attribute.String("error.type", strconv.Itoa(statusCode)))
	}

	// The convention defines this instrument in milliseconds.
	m.durationHistogram.Record(ctx, float64(duration.Nanoseconds())/1e6, metric.WithAttributes(attrs...))
}
