package requests_test

import (
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenValidatingAPaymentRequestTestingSuite struct{}

func (suite *WhenValidatingAPaymentRequestTestingSuite) assertInvalid(t *testing.T, err error, reason string) {
	var invalidPayment *apperrors.InvalidPaymentError
	require.ErrorAs(t, err, &invalidPayment)
	assert.Equal(t, reason, invalidPayment.Reason)
}

func TestWhenValidatingAPaymentRequest(t *testing.T) {
	t.Parallel()

	suite := &WhenValidatingAPaymentRequestTestingSuite{}

	t.Run("Given the payment payload is missing", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload = nil

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the payment requirements are missing", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements = nil

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the request is not x402 version 2", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.X402Version = 1

		t.Run("Should refuse the version", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidX402Version.String())
		})
	})

	t.Run("Given the payload is not x402 version 2", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.X402Version = 1

		t.Run("Should refuse the version", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidX402Version.String())
		})
	})

	t.Run("Given the requirements scheme is not exact", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Scheme = "upto"

		t.Run("Should refuse the scheme as unsupported", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonUnsupportedScheme.String())
		})
	})

	t.Run("Given the payload was signed for another scheme", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Accepted.Scheme = "upto"

		t.Run("Should refuse the scheme", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidScheme.String())
		})
	})

	t.Run("Given the network is not a CAIP-2 eip155 identifier", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Network = "polygon"
		request.PaymentPayload.Accepted.Network = "polygon"

		t.Run("Should refuse the network", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidNetwork.String())
		})
	})

	t.Run("Given the network is not written in its canonical form", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Network = "eip155:080002"
		request.PaymentPayload.Accepted.Network = "eip155:080002"

		t.Run("Should refuse the network", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidNetwork.String())
		})
	})

	t.Run("Given the payload was signed for another network", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Accepted.Network = "eip155:137"

		t.Run("Should refuse the network", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidNetwork.String())
		})
	})

	t.Run("Given the requirements ask for another asset transfer method", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Extra["assetTransferMethod"] = "permit2"

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the asset is not an address", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Asset = "usdc"

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given payTo is not an address", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.PayTo = "agent"

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the amount is zero", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Amount = "0"
		request.PaymentPayload.Accepted.Amount = "0"

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the amount is not a decimal integer", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentRequirements.Amount = "1e3"
		request.PaymentPayload.Accepted.Amount = "1e3"

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the accepted requirements differ from the submitted ones", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Accepted.Amount = "1"

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the authorization is missing", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization = nil

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the authorization from is not an address", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization.From = "payer"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the authorization validBefore is not a decimal integer", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization.ValidBefore = "soon"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the authorization nonce is not 32 bytes", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization.Nonce = "0x01"

		t.Run("Should refuse it as an invalid payload", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPayload.String())
		})
	})

	t.Run("Given the signature is not 65 bytes", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Signature = "0xdeadbeef"

		t.Run("Should refuse the signature", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidSignature.String())
		})
	})

	t.Run("Given the authorization pays someone other than payTo", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization.To = common_tests.RandomEthAddress().Hex()

		t.Run("Should refuse the recipient", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidRecipient.String())
		})
	})

	t.Run("Given the authorization value differs from the amount", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()
		request.PaymentPayload.Payload.Authorization.Value = "999"

		t.Run("Should refuse the value", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidAuthorizationValue.String())
		})
	})

	t.Run("Given a legacy payload with a top-level scheme and network", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithLegacyPayload().Build()

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			assert.NoError(t, err)
		})
	})

	t.Run("Given a valid payment", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().Build()

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			err := request.ValidateAndSanitize()

			assert.NoError(t, err)
		})
	})
}
