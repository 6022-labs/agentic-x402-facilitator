package services_test

import (
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenVerifyingAnEip3009SignatureTestingSuite struct {
	token *x402_facilitator_settings.TokenConfig
	sut   *x402_facilitator_services.Eip3009AssetTransferMethodStrategy
}

func WhenVerifyingAnEip3009SignatureBeforeEach() *WhenVerifyingAnEip3009SignatureTestingSuite {
	return &WhenVerifyingAnEip3009SignatureTestingSuite{
		token: &x402_facilitator_settings.TokenConfig{Eip712Name: "USDC", Eip712Version: "2"},
		sut:   x402_facilitator_services.NewEip3009AssetTransferMethodStrategy(zap.NewNop(), nil),
	}
}

func TestWhenVerifyingAnEip3009Signature(t *testing.T) {
	t.Parallel()

	t.Run("Given the payer signed under another EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithEip712Domain("USD Coin", "2").BuildPayment()

		t.Run("Should refuse the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnEip3009SignatureBeforeEach()

			err := suite.sut.VerifySignature(payment, suite.token)

			var invalidPayment *apperrors.InvalidPaymentError
			require.ErrorAs(t, err, &invalidPayment)
			assert.Equal(t, enums.PaymentReasonInvalidSignature.String(), invalidPayment.Reason)
		})
	})

	t.Run("Given the payer signed with a v of 0 or 1", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithRawRecoveryId().BuildPayment()

		t.Run("Should accept the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnEip3009SignatureBeforeEach()

			err := suite.sut.VerifySignature(payment, suite.token)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the payer signed the payment", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		t.Run("Should accept the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnEip3009SignatureBeforeEach()

			err := suite.sut.VerifySignature(payment, suite.token)

			assert.NoError(t, err)
		})
	})
}
