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
	mockStrategies *interfaces_mocks.MockAssetTransferMethodStrategiesInterface
	mockStrategy   *interfaces_mocks.MockAssetTransferMethodStrategyInterface
	sut            *x402_facilitator_services.PaymentVerifier
}

// WhenVerifyingAPaymentAuthorizationBeforeEach configures asset for both methods.
func WhenVerifyingAPaymentAuthorizationBeforeEach(
	t *testing.T,
	asset common.Address,
) *WhenVerifyingAPaymentAuthorizationTestingSuite {
	mockController := gomock.NewController(t)
	mockStrategies := interfaces_mocks.NewMockAssetTransferMethodStrategiesInterface(mockController)

	return &WhenVerifyingAPaymentAuthorizationTestingSuite{
		mockStrategies: mockStrategies,
		mockStrategy:   interfaces_mocks.NewMockAssetTransferMethodStrategyInterface(mockController),
		sut: x402_facilitator_services.NewPaymentVerifier(
			&x402_facilitator_settings.TokensSettings{
				80002: {asset: {
					AssetTransferMethods: []enums.AssetTransferMethod{
						enums.AssetTransferMethodEip3009,
						enums.AssetTransferMethodPermit2,
					},
					Eip712Name:    "USDC",
					Eip712Version: "2",
				}},
			},
			interfaces_mocks.NewMockPayToEligibilityCheckerInterface(mockController),
			mockStrategies,
			interfaces_mocks.NewMockTokenRequesterInterface(mockController),
			interfaces_mocks.NewMockTransferSubmitterInterface(mockController),
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

	t.Run("Given no strategy settles the payment's transfer method", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithPermit2().BuildPayment()

		initSuite := func(suite *WhenVerifyingAPaymentAuthorizationTestingSuite) {
			suite.mockStrategies.EXPECT().For(enums.AssetTransferMethodPermit2).Return(nil, false)
		}

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentAuthorizationBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.VerifyAuthorization(payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the strategy refuses the signature", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()
		signatureErr := apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidSignature, "signature mismatch")

		initSuite := func(suite *WhenVerifyingAPaymentAuthorizationTestingSuite) {
			suite.mockStrategies.EXPECT().For(enums.AssetTransferMethodEip3009).Return(suite.mockStrategy, true)
			suite.mockStrategy.EXPECT().VerifySignature(payment, gomock.Any()).Return(signatureErr)
		}

		t.Run("Should return the strategy's refusal", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentAuthorizationBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.VerifyAuthorization(payment)

			assert.Same(t, signatureErr, err)
		})
	})

	t.Run("Given an expired authorization signed by its payer", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithValidBefore(time.Now().Unix() - 60).BuildPayment()

		initSuite := func(suite *WhenVerifyingAPaymentAuthorizationTestingSuite) {
			suite.mockStrategies.EXPECT().For(enums.AssetTransferMethodEip3009).Return(suite.mockStrategy, true)
			suite.mockStrategy.EXPECT().VerifySignature(payment, gomock.Any()).Return(nil)
		}

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentAuthorizationBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.VerifyAuthorization(payment)

			assert.NoError(t, err)
		})
	})
}
