package metrics

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	common_metrics "github.com/6022-labs/agentic-x402-facilitator/src/common/metrics"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const (
	meterName = "agentic_x402_facilitator"

	operationVerify = "verify"
	operationSettle = "settle"

	unknownLabel     = "unknown"
	unsupportedLabel = "unsupported"
	otherLabel       = "other"
)

type FacilitatorMetrics struct {
	tokensSettings       *x402_facilitator_settings.TokensSettings
	accessSettings       *x402_facilitator_settings.AccessSettings
	verificationDuration metric.Float64Histogram
	settlementDuration   metric.Float64Histogram
	paymentAmount        metric.Float64Counter
}

func NewFacilitatorMetrics(
	tokensSettings *x402_facilitator_settings.TokensSettings,
	accessSettings *x402_facilitator_settings.AccessSettings,
) *FacilitatorMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	verificationDuration, err := meter.Float64Histogram(
		"x402.facilitator.verification.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LatencySecondsBuckets...),
		metric.WithDescription("Duration of x402 payment verifications"),
	)
	if err != nil {
		verificationDuration, _ = noopMeter.Float64Histogram("x402.facilitator.verification.duration")
	}

	settlementDuration, err := meter.Float64Histogram(
		"x402.facilitator.settlement.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LongLatencySecondsBuckets...),
		metric.WithDescription("Duration of x402 payment settlements, receipt wait included"),
	)
	if err != nil {
		settlementDuration, _ = noopMeter.Float64Histogram("x402.facilitator.settlement.duration")
	}

	paymentAmount, err := meter.Float64Counter(
		"x402.facilitator.payment.amount",
		metric.WithUnit("{token_base_unit}"),
		metric.WithDescription("Payment amounts, in token base units"),
	)
	if err != nil {
		paymentAmount, _ = noopMeter.Float64Counter("x402.facilitator.payment.amount")
	}

	return &FacilitatorMetrics{
		tokensSettings:       tokensSettings,
		accessSettings:       accessSettings,
		verificationDuration: verificationDuration,
		settlementDuration:   settlementDuration,
		paymentAmount:        paymentAmount,
	}
}

func (m *FacilitatorMetrics) RecordVerification(
	ctx context.Context,
	payment *models.Payment,
	verified bool,
	outcome string,
	duration time.Duration,
	err error,
) {
	attrs := m.Attributes(payment, verified, outcome, err)
	m.verificationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	m.recordAmount(ctx, payment, verified, operationVerify, attrs)
}

func (m *FacilitatorMetrics) RecordSettlement(
	ctx context.Context,
	payment *models.Payment,
	verified bool,
	outcome string,
	duration time.Duration,
	err error,
) {
	attrs := m.Attributes(payment, verified, outcome, err)
	m.settlementDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	m.recordAmount(ctx, payment, verified, operationSettle, attrs)
}

// Attributes labels only configured tokens and verified agent payees, since callers choose what they send.
func (m *FacilitatorMetrics) Attributes(
	payment *models.Payment,
	verified bool,
	outcome string,
	err error,
) []attribute.KeyValue {
	network, asset := unknownLabel, unknownLabel
	if payment != nil && m.tokensSettings.HasChain(payment.ChainId) {
		network = helpers.FormatCaip2(payment.ChainId)
		asset = unsupportedLabel
		if m.tokensSettings.Get(payment.ChainId, payment.Asset) != nil {
			asset = strings.ToLower(payment.Asset.Hex())
		}
	}
	payTo := otherLabel
	if payment != nil && verified && m.accessSettings.AgentsOnly {
		payTo = strings.ToLower(payment.PayTo.Hex())
	}
	method := unknownLabel
	if payment != nil {
		method = payment.Method.String()
	}
	attrs := []attribute.KeyValue{
		attribute.String("x402.network", network),
		attribute.String("x402.asset", asset),
		attribute.String("x402.asset_transfer_method", method),
		attribute.String("x402.pay_to", payTo),
	}

	var unavailable *apperrors.UnavailableError
	switch {
	case errors.As(err, &unavailable):
		attrs = append(attrs, attribute.String("x402.outcome", "unavailable"))
	case err != nil:
		attrs = append(attrs, attribute.String("x402.outcome", "error"))
	default:
		attrs = append(attrs, attribute.String("x402.outcome", outcome))
	}
	// error.type is conditionally required, and only on a failed request.
	if err != nil {
		attrs = append(attrs, attribute.String("error.type", common_metrics.ErrorType(err)))
	}
	return attrs
}

// recordAmount skips unverified payments: anyone can ask for any amount.
func (m *FacilitatorMetrics) recordAmount(
	ctx context.Context,
	payment *models.Payment,
	verified bool,
	operation string,
	attrs []attribute.KeyValue,
) {
	if payment == nil || !verified {
		return
	}
	amount, _ := new(big.Float).SetInt(payment.Amount).Float64()
	m.paymentAmount.Add(
		ctx,
		amount,
		metric.WithAttributes(append(attrs, attribute.String("x402.operation", operation))...),
	)
}
