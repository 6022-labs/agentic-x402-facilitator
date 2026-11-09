package services_test

import (
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenReadingAPermit2PaymentTestingSuite struct {
	sut *x402_facilitator_services.Permit2AssetTransferMethodStrategy
}

func WhenReadingAPermit2PaymentBeforeEach() *WhenReadingAPermit2PaymentTestingSuite {
	return &WhenReadingAPermit2PaymentTestingSuite{
		sut: x402_facilitator_services.NewPermit2AssetTransferMethodStrategy(zap.NewNop(), nil, nil),
	}
}

func (suite *WhenReadingAPermit2PaymentTestingSuite) assertInvalid(t *testing.T, err error, reason string) {
	var invalidPayment *apperrors.InvalidPaymentError
	require.ErrorAs(t, err, &invalidPayment)
	assert.Equal(t, reason, invalidPayment.Reason)
}

func TestWhenReadingAPermit2Payment(t *testing.T) {
	t.Parallel()

	t.Run("Given the permit2 authorization is missing", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Permit2Authorization = nil

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the spender is not an address", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Permit2Authorization.Spender = "proxy"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the nonce is not a decimal integer", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Permit2Authorization.Nonce = "0x01"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the signature is not 65 bytes", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Signature = "0xdeadbeef"

		t.Run("Should refuse the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPermit2Signature.String())
		})
	})

	t.Run("Given the spender is not x402ExactPermit2Proxy", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Permit2Authorization.Spender = common_tests.RandomEthAddress().Hex()

		t.Run("Should refuse the spender", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPermit2Spender.String())
		})
	})

	t.Run("Given the witness pays someone other than payTo", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Permit2Authorization.Witness.To = common_tests.RandomEthAddress().Hex()

		t.Run("Should refuse the recipient", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPermit2RecipientMismatch.String())
		})
	})

	t.Run("Given the permitted token is not the asset", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Permit2Authorization.Permitted.Token = common_tests.RandomEthAddress().Hex()

		t.Run("Should refuse the token", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonPermit2TokenMismatch.String())
		})
	})

	t.Run("Given the permitted amount differs from the amount", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().Build()
		request.PaymentPayload.Payload.Permit2Authorization.Permitted.Amount = "999"

		t.Run("Should refuse the amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonPermit2AmountMismatch.String())
		})
	})

	t.Run("Given a valid Permit2 payment", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder().WithPermit2()
		request := builder.Build()

		t.Run("Should read the payment the payer signed", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAPermit2PaymentBeforeEach()

			payment, err := suite.sut.Payment(request)

			require.NoError(t, err)
			assert.Equal(t, builder.BuildPayment(), payment)
		})
	})
}
