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

type WhenReadingAnEip3009PaymentTestingSuite struct {
	sut *x402_facilitator_services.Eip3009AssetTransferMethodStrategy
}

func WhenReadingAnEip3009PaymentBeforeEach() *WhenReadingAnEip3009PaymentTestingSuite {
	return &WhenReadingAnEip3009PaymentTestingSuite{
		sut: x402_facilitator_services.NewEip3009AssetTransferMethodStrategy(zap.NewNop(), nil),
	}
}

func (suite *WhenReadingAnEip3009PaymentTestingSuite) assertInvalid(t *testing.T, err error, reason string) {
	var invalidPayment *apperrors.InvalidPaymentError
	require.ErrorAs(t, err, &invalidPayment)
	assert.Equal(t, reason, invalidPayment.Reason)
}

func TestWhenReadingAnEip3009Payment(t *testing.T) {
	t.Parallel()

	t.Run("Given the authorization is missing", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Eip3009Authorization = nil

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the authorization from is not an address", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Eip3009Authorization.From = "payer"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the authorization validBefore is not a decimal integer", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Eip3009Authorization.ValidBefore = "soon"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the authorization nonce is not 32 bytes", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Eip3009Authorization.Nonce = "0x01"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the signature is not 65 bytes", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Signature = "0xdeadbeef"

		t.Run("Should refuse the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidSignature.String())
		})
	})

	t.Run("Given the authorization pays someone other than payTo", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Eip3009Authorization.To = common_tests.RandomEthAddress().Hex()

		t.Run("Should refuse the recipient", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidRecipient.String())
		})
	})

	t.Run("Given the authorization value differs from the amount", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Eip3009Authorization.Value = "999"

		t.Run("Should refuse the value", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			_, err := suite.sut.Payment(request)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidAuthorizationValue.String())
		})
	})

	t.Run("Given a valid EIP-3009 payment", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.Build()

		t.Run("Should read the payment the payer signed", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnEip3009PaymentBeforeEach()

			payment, err := suite.sut.Payment(request)

			require.NoError(t, err)
			assert.Equal(t, builder.BuildPayment(), payment)
		})
	})
}
