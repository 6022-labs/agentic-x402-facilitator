package metrics_test

import (
	"errors"
	"strings"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/metrics"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

type WhenLabelingAPaymentMeasurementTestingSuite struct {
	sut *metrics.FacilitatorMetrics
}

func WhenLabelingAPaymentMeasurementBeforeEach(
	asset common.Address,
	agentsOnly bool,
) *WhenLabelingAPaymentMeasurementTestingSuite {
	return &WhenLabelingAPaymentMeasurementTestingSuite{
		sut: metrics.NewFacilitatorMetrics(
			&x402_facilitator_settings.TokensSettings{
				80002: {asset: {Eip712Name: "USDC", Eip712Version: "2"}},
			},
			&x402_facilitator_settings.AccessSettings{AgentsOnly: agentsOnly},
		),
	}
}

func (suite *WhenLabelingAPaymentMeasurementTestingSuite) labels(attrs []attribute.KeyValue) map[string]string {
	return lo.SliceToMap(attrs, func(attr attribute.KeyValue) (string, string) {
		return string(attr.Key), attr.Value.AsString()
	})
}

func TestWhenLabelingAPaymentMeasurement(t *testing.T) {
	t.Parallel()

	asset := common_tests.RandomEthAddress()

	t.Run("Given a request that never parsed into a payment", func(t *testing.T) {
		t.Parallel()

		t.Run("Should label network and asset unknown", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(
				suite.sut.Attributes(nil, false, enums.PaymentReasonInvalidX402Version.String(), nil),
			)

			assert.Equal(t, "unknown", labels["x402.network"])
			assert.Equal(t, "unknown", labels["x402.asset"])
		})
	})

	t.Run("Given a payment on a network the facilitator does not serve", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithChainId(999).BuildPayment()

		t.Run("Should label network and asset unknown", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(
				suite.sut.Attributes(payment, false, enums.PaymentReasonInvalidNetwork.String(), nil),
			)

			assert.Equal(t, "unknown", labels["x402.network"])
			assert.Equal(t, "unknown", labels["x402.asset"])
		})
	})

	t.Run("Given a payment in a token the facilitator does not settle", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		t.Run("Should label its network and the asset unsupported", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(
				suite.sut.Attributes(payment, false, enums.PaymentReasonInvalidPaymentRequirements.String(), nil),
			)

			assert.Equal(t, "eip155:80002", labels["x402.network"])
			assert.Equal(t, "unsupported", labels["x402.asset"])
		})
	})

	t.Run("Given a payment that failed verification", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithAsset(asset).BuildPayment()

		t.Run("Should label the payee other", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(
				suite.sut.Attributes(payment, false, enums.PaymentReasonInsufficientFunds.String(), nil),
			)

			assert.Equal(t, "other", labels["x402.pay_to"])
		})
	})

	t.Run("Given a verified payment while access is open to every payee", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithAsset(asset).BuildPayment()

		t.Run("Should label the payee other", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, false)

			labels := suite.labels(suite.sut.Attributes(payment, true, "valid", nil))

			assert.Equal(t, "other", labels["x402.pay_to"])
		})
	})

	t.Run("Given an unavailable dependency", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithAsset(asset).BuildPayment()
		err := apperrors.NewUnavailableError("chain_unavailable", "the chain RPC is unavailable", nil)

		t.Run("Should label the outcome unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(suite.sut.Attributes(payment, false, "", err))

			assert.Equal(t, "unavailable", labels["x402.outcome"])
		})

		t.Run("Should add an error type", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(suite.sut.Attributes(payment, false, "", err))

			assert.Equal(t, "_OTHER", labels["error.type"])
		})
	})

	t.Run("Given an unexpected error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("boom")

		t.Run("Should label the outcome error", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(suite.sut.Attributes(nil, false, "", err))

			assert.Equal(t, "error", labels["x402.outcome"])
		})
	})

	t.Run("Given a settled payment in a configured token", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithAsset(asset).BuildPayment()

		t.Run("Should label its network, token and outcome", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(suite.sut.Attributes(payment, true, "settled", nil))

			assert.Equal(t, "eip155:80002", labels["x402.network"])
			assert.Equal(t, strings.ToLower(asset.Hex()), labels["x402.asset"])
			assert.Equal(t, "settled", labels["x402.outcome"])
		})

		t.Run("Should label its asset transfer method", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(suite.sut.Attributes(payment, true, "settled", nil))

			assert.Equal(t, "eip3009", labels["x402.asset_transfer_method"])
		})

		t.Run("Should label its payee", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(suite.sut.Attributes(payment, true, "settled", nil))

			assert.Equal(t, strings.ToLower(payment.PayTo.Hex()), labels["x402.pay_to"])
		})

		t.Run("Should add no error type", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabelingAPaymentMeasurementBeforeEach(asset, true)

			labels := suite.labels(suite.sut.Attributes(payment, true, "settled", nil))

			assert.NotContains(t, labels, "error.type")
		})
	})
}
