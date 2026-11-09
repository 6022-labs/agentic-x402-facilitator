package services_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenVerifyingAnEip3009PaymentOnChainTestingSuite struct {
	mockTokenRequester *interfaces_mocks.MockTokenRequesterInterface
	sut                *x402_facilitator_services.Eip3009AssetTransferMethodStrategy
}

func WhenVerifyingAnEip3009PaymentOnChainBeforeEach(t *testing.T) *WhenVerifyingAnEip3009PaymentOnChainTestingSuite {
	mockTokenRequester := interfaces_mocks.NewMockTokenRequesterInterface(gomock.NewController(t))
	return &WhenVerifyingAnEip3009PaymentOnChainTestingSuite{
		mockTokenRequester: mockTokenRequester,
		sut: x402_facilitator_services.NewEip3009AssetTransferMethodStrategy(
			zap.NewNop(),
			mockTokenRequester,
		),
	}
}

func TestWhenVerifyingAnEip3009PaymentOnChain(t *testing.T) {
	t.Parallel()

	payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

	t.Run("Given the authorization state cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenVerifyingAnEip3009PaymentOnChainTestingSuite) {
			suite.mockTokenRequester.EXPECT().
				AuthorizationState(
					gomock.Any(),
					uint64(80002),
					payment.Asset,
					payment.Authorization.From,
					payment.Authorization.Nonce,
				).
				Return(false, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnEip3009PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the authorization nonce is already used", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAnEip3009PaymentOnChainTestingSuite) {
			suite.mockTokenRequester.EXPECT().
				AuthorizationState(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, nil)
		}

		t.Run("Should refuse the transaction state", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnEip3009PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			var invalidPayment *apperrors.InvalidPaymentError
			require.ErrorAs(t, err, &invalidPayment)
			assert.Equal(t, enums.PaymentReasonInvalidTransactionState.String(), invalidPayment.Reason)
		})
	})

	t.Run("Given the authorization nonce is unused", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAnEip3009PaymentOnChainTestingSuite) {
			suite.mockTokenRequester.EXPECT().
				AuthorizationState(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil)
		}

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnEip3009PaymentOnChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.VerifyOnChain(context.Background(), payment)

			assert.NoError(t, err)
		})
	})
}
