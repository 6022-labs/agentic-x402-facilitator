package services_test

import (
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenVerifyingAPaymentAuthorizationTestingSuite struct {
	sut *x402_facilitator_services.PaymentVerifier
}

func WhenVerifyingAPaymentAuthorizationBeforeEach(
	t *testing.T,
	asset common.Address,
) *WhenVerifyingAPaymentAuthorizationTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenVerifyingAPaymentAuthorizationTestingSuite{
		sut: x402_facilitator_services.NewPaymentVerifier(
			&x402_facilitator_settings.TokensSettings{
				80002: {asset: {Eip712Name: "USDC", Eip712Version: "2"}},
			},
			interfaces_mocks.NewMockPayToEligibilityCheckerInterface(mockController),
			interfaces_mocks.NewMockEip3009TokenRequesterInterface(mockController),
		),
	}
}

func (suite *WhenVerifyingAPaymentAuthorizationTestingSuite) assertInvalid(t *testing.T, err error, reason string) {
	var invalidPayment *apperrors.InvalidPaymentError
	require.ErrorAs(t, err, &invalidPayment)
	assert.Equal(t, reason, invalidPayment.Reason)
}

func TestWhenVerifyingAPaymentAuthorization(t *testing.T) {
	t.Parallel()

	asset := common_tests.RandomEthAddress()
	builder := func() *model_builders.PaymentRequestBuilder {
		return model_builders.NewPaymentRequestBuilder().WithAsset(asset)
	}

	t.Run("Given the network has no configured token", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithChainId(137).BuildPayment()

		t.Run("Should refuse the network", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentAuthorizationBeforeEach(t, asset)

			err := suite.sut.VerifyAuthorization(payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidNetwork.String())
		})
	})

	t.Run("Given the asset is not a configured token", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentAuthorizationBeforeEach(t, asset)

			err := suite.sut.VerifyAuthorization(payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given a payee other than the one signed", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()
		payment.Authorization.To = common_tests.RandomEthAddress()

		t.Run("Should refuse the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentAuthorizationBeforeEach(t, asset)

			err := suite.sut.VerifyAuthorization(payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidSignature.String())
		})
	})

	t.Run("Given an expired authorization signed by its payer", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithValidBefore(time.Now().Unix() - 60).BuildPayment()

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentAuthorizationBeforeEach(t, asset)

			err := suite.sut.VerifyAuthorization(payment)

			assert.NoError(t, err)
		})
	})
}
