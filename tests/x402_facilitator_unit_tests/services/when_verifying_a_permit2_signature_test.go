package services_test

import (
	"math/big"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenVerifyingAPermit2SignatureTestingSuite struct {
	sut *x402_facilitator_services.Permit2AssetTransferMethodStrategy
}

func WhenVerifyingAPermit2SignatureBeforeEach() *WhenVerifyingAPermit2SignatureTestingSuite {
	return &WhenVerifyingAPermit2SignatureTestingSuite{
		sut: x402_facilitator_services.NewPermit2AssetTransferMethodStrategy(zap.NewNop(), nil, nil),
	}
}

func TestWhenVerifyingAPermit2Signature(t *testing.T) {
	t.Parallel()

	t.Run("Given the signed terms were changed", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithPermit2().BuildPayment()
		payment.Authorization.ValidBefore = new(big.Int).Add(payment.Authorization.ValidBefore, big.NewInt(1))

		t.Run("Should refuse the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPermit2SignatureBeforeEach()

			err := suite.sut.VerifySignature(payment, nil)

			var invalidPayment *apperrors.InvalidPaymentError
			require.ErrorAs(t, err, &invalidPayment)
			assert.Equal(t, enums.PaymentReasonInvalidPermit2Signature.String(), invalidPayment.Reason)
		})
	})

	t.Run("Given the payer signed the payment", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithPermit2().BuildPayment()

		t.Run("Should accept the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPermit2SignatureBeforeEach()

			err := suite.sut.VerifySignature(payment, nil)

			assert.NoError(t, err)
		})
	})
}
