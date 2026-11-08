package metrics

import (
	"context"
	"math/big"
	"strconv"

	blockchain_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services/interfaces"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
)

var weiPerNativeToken = new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

// RelayerMetrics reads the relayer balance at each collection, so telemetry off costs no RPC call.
type RelayerMetrics struct {
	gasUsed  metric.Int64Counter
	gasSpent metric.Float64Counter
}

func NewRelayerMetrics(
	relayerBalanceReader blockchain_interfaces.RelayerBalanceReaderInterface,
	logger *zap.Logger,
) *RelayerMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	gasUsed, err := meter.Int64Counter(
		"x402.facilitator.relayer.gas.used",
		metric.WithUnit("{gas}"),
		metric.WithDescription("Gas units the relayer's mined settlements used"),
	)
	if err != nil {
		gasUsed, _ = noopMeter.Int64Counter("x402.facilitator.relayer.gas.used")
	}

	gasSpent, err := meter.Float64Counter(
		"x402.facilitator.relayer.gas.spent",
		metric.WithUnit("{native_token}"),
		metric.WithDescription("Native token the relayer paid for its mined settlements"),
	)
	if err != nil {
		gasSpent, _ = noopMeter.Float64Counter("x402.facilitator.relayer.gas.spent")
	}

	relayerMetrics := &RelayerMetrics{gasUsed: gasUsed, gasSpent: gasSpent}

	_, err = meter.Float64ObservableGauge(
		"x402.facilitator.relayer.balance",
		metric.WithUnit("{native_token}"),
		metric.WithDescription("Native token balance of the relayer wallet"),
		metric.WithFloat64Callback(func(ctx context.Context, observer metric.Float64Observer) error {
			for chainId, balance := range relayerBalanceReader.Balances(ctx) {
				observer.Observe(
					relayerMetrics.toNativeToken(balance),
					metric.WithAttributes(relayerMetrics.chainAttribute(chainId)),
				)
			}
			return nil
		}),
	)
	if err != nil {
		logger.Warn("relayer balance gauge not registered", zap.Error(err))
	}

	return relayerMetrics
}

func (m *RelayerMetrics) RecordGasSpent(
	ctx context.Context,
	chainId uint64,
	gasUsed uint64,
	costWei *big.Int,
	status string,
) {
	attrs := metric.WithAttributes(m.chainAttribute(chainId), attribute.String("x402.settlement.status", status))
	m.gasUsed.Add(ctx, int64(gasUsed), attrs)
	m.gasSpent.Add(ctx, m.toNativeToken(costWei), attrs)
}

func (m *RelayerMetrics) toNativeToken(wei *big.Int) float64 {
	value, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), weiPerNativeToken).Float64()
	return value
}

func (m *RelayerMetrics) chainAttribute(chainId uint64) attribute.KeyValue {
	return attribute.String("chain.id", strconv.FormatUint(chainId, 10))
}
