package services_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenVerifyingAPermit2PaymentOnChainTestingSuite struct {
	mockTokenRequester   *interfaces_mocks.MockTokenRequesterInterface
	mockPermit2Requester *interfaces_mocks.MockPermit2RequesterInterface
	sut                  *x402_facilitator_services.Permit2AssetTransferMethodStrategy
}

func WhenVerifyingAPermit2PaymentOnChainBeforeEach(t *testing.T) *WhenVerifyingAPermit2PaymentOnChainTestingSuite {
	mockController := gomock.NewController(t)
	mockTokenRequester := interfaces_mocks.NewMockTokenRequesterInterface(mockController)
	mockPermit2Requester := interfaces_mocks.NewMockPermit2RequesterInterface(mockController)
	return &WhenVerifyingAPermit2PaymentOnChainTestingSuite{
		mockTokenRequester:   mockTokenRequester,
		mockPermit2Requester: mockPermit2Requester,
		sut: x402_facilitator_services.NewPermit2AssetTransferMethodStrategy(
			zap.NewNop(),
			mockTokenRequester,
			mockPermit2Requester,
		),
	}
}

func (suite *WhenVerifyingAPermit2PaymentOnChainTestingSuite) assertInvalid(t *testing.T, err error, reason string) {
	var invalidPayment *apperrors.InvalidPaymentError
	require.ErrorAs(t, err, &invalidPayment)
	assert.Equal(t, reason, invalidPayment.Reason)
}

func TestWhenVerifyingAPermit2PaymentOnChain(t *testing.T) {
	t.Parallel()

	payment := model_builders.NewPaymentRequestBuilder().WithPermit2().BuildPayment()

	t.Run("Given the proxy code cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenVerifyingAPermit2PaymentOnChainTestingSuite) {
			suite.mockPermit2Requester.EXPECT().ProxyDeployed(gomock.Any(), uint64(80002)).Return(false, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPermit2PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the proxy is not deployed on the network", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAPermit2PaymentOnChainTestingSuite) {
			suite.mockPermit2Requester.EXPECT().ProxyDeployed(gomock.Any(), gomock.Any()).Return(false, nil)
		}

		t.Run("Should refuse the network", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPermit2PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonPermit2ProxyNotDeployed.String())
		})
	})

	t.Run("Given the nonce is already used", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAPermit2PaymentOnChainTestingSuite) {
			suite.mockPermit2Requester.EXPECT().ProxyDeployed(gomock.Any(), gomock.Any()).Return(true, nil)
			suite.mockPermit2Requester.EXPECT().
				NonceUsed(gomock.Any(), uint64(80002), payment.Authorization.From, payment.Authorization.NonceInt()).
				Return(true, nil)
		}

		t.Run("Should refuse the transaction state", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPermit2PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidTransactionState.String())
		})
	})

	t.Run("Given the payer has not approved Permit2 for the amount", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAPermit2PaymentOnChainTestingSuite) {
			suite.mockPermit2Requester.EXPECT().ProxyDeployed(gomock.Any(), gomock.Any()).Return(true, nil)
			suite.mockPermit2Requester.EXPECT().
				NonceUsed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil)
			suite.mockTokenRequester.EXPECT().
				Allowance(
					gomock.Any(),
					uint64(80002),
					payment.Asset,
					payment.Authorization.From,
					common.HexToAddress(helpers.Permit2Address),
				).
				Return(big.NewInt(1), nil)
		}

		t.Run("Should ask for the Permit2 allowance", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPermit2PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonPermit2AllowanceRequired.String())
		})
	})

	t.Run("Given the proxy is deployed, the nonce unused and Permit2 approved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAPermit2PaymentOnChainTestingSuite) {
			suite.mockPermit2Requester.EXPECT().ProxyDeployed(gomock.Any(), gomock.Any()).Return(true, nil)
			suite.mockPermit2Requester.EXPECT().
				NonceUsed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil)
			suite.mockTokenRequester.EXPECT().
				Allowance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(payment.Amount, nil)
		}

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPermit2PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			assert.NoError(t, err)
		})
	})
}
