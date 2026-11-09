package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReadingARememberedSettlementTestingSuite struct {
	mockCacheStorage *interfaces_mocks.MockCacheStorageInterface
	sut              *x402_facilitator_services.SettlementGuard
}

func WhenReadingARememberedSettlementBeforeEach(t *testing.T) *WhenReadingARememberedSettlementTestingSuite {
	mockController := gomock.NewController(t)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)

	return &WhenReadingARememberedSettlementTestingSuite{
		mockCacheStorage: mockCacheStorage,
		sut: x402_facilitator_services.NewSettlementGuard(
			interfaces_mocks.NewMockLockStorageInterface(mockController),
			mockCacheStorage,
			&x402_facilitator_settings.SettlementSettings{LockTtl: 10 * time.Minute},
			zap.NewNop(),
		),
	}
}

func TestWhenReadingARememberedSettlement(t *testing.T) {
	t.Parallel()

	payment := model_builders.NewPaymentRequestBuilder().BuildPayment()
	txKey := "x402:settlement:" + payment.SettlementKey() + ":tx"

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("redis down")

		initSuite := func(suite *WhenReadingARememberedSettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().ReadString(gomock.Any(), txKey).Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingARememberedSettlementBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.RememberedTransaction(context.Background(), payment)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given no transaction is remembered", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReadingARememberedSettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().ReadString(gomock.Any(), txKey).Return(nil, nil)
		}

		t.Run("Should return nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingARememberedSettlementBeforeEach(t)
			initSuite(suite)

			txHash, err := suite.sut.RememberedTransaction(context.Background(), payment)

			assert.NoError(t, err)
			assert.Nil(t, txHash)
		})
	})

	t.Run("Given a transaction remembered for other terms under the same nonce", func(t *testing.T) {
		t.Parallel()

		other := model_builders.NewPaymentRequestBuilder().WithAmount("2").BuildPayment()
		stored := common.HexToHash("0xabc").Hex() + " " + other.Terms()

		initSuite := func(suite *WhenReadingARememberedSettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().ReadString(gomock.Any(), txKey).Return(&stored, nil)
		}

		t.Run("Should refuse the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingARememberedSettlementBeforeEach(t)
			initSuite(suite)

			txHash, err := suite.sut.RememberedTransaction(context.Background(), payment)

			var invalidPayment *apperrors.InvalidPaymentError
			require.ErrorAs(t, err, &invalidPayment)
			assert.Equal(t, enums.PaymentReasonInvalidTransactionState.String(), invalidPayment.Reason)
			assert.Nil(t, txHash)
		})
	})

	t.Run("Given a transaction is remembered", func(t *testing.T) {
		t.Parallel()

		stored := common.HexToHash("0xabc").Hex() + " " + payment.Terms()

		initSuite := func(suite *WhenReadingARememberedSettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().ReadString(gomock.Any(), txKey).Return(&stored, nil)
		}

		t.Run("Should return its hash", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingARememberedSettlementBeforeEach(t)
			initSuite(suite)

			txHash, err := suite.sut.RememberedTransaction(context.Background(), payment)

			require.NoError(t, err)
			require.NotNil(t, txHash)
			assert.Equal(t, common.HexToHash("0xabc"), *txHash)
		})
	})
}
