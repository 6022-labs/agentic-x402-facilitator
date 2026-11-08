package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRememberingASettlementTestingSuite struct {
	mockCacheStorage *interfaces_mocks.MockCacheStorageInterface
	sut              *x402_facilitator_services.SettlementGuard
}

func WhenRememberingASettlementBeforeEach(t *testing.T) *WhenRememberingASettlementTestingSuite {
	mockController := gomock.NewController(t)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)

	return &WhenRememberingASettlementTestingSuite{
		mockCacheStorage: mockCacheStorage,
		sut: x402_facilitator_services.NewSettlementGuard(
			interfaces_mocks.NewMockLockStorageInterface(mockController),
			mockCacheStorage,
			&x402_facilitator_settings.SettlementSettings{LockTtl: 10 * time.Minute},
			zap.NewNop(),
		),
	}
}

func TestWhenRememberingASettlement(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0xabc")
	paymentExpiringIn := func(validity time.Duration) *models.Payment {
		return model_builders.NewPaymentRequestBuilder().WithValidBefore(time.Now().Add(validity).Unix()).BuildPayment()
	}
	txKey := func(payment *models.Payment) string {
		return "x402:settlement:" + payment.SettlementKey() + ":tx"
	}
	stored := func(payment *models.Payment) string {
		return txHash.Hex() + " " + payment.Authorization.Terms()
	}

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("redis down")
		payment := paymentExpiringIn(time.Hour)

		initSuite := func(suite *WhenRememberingASettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().
				Store(gomock.Any(), txKey(payment), stored(payment), gomock.Any()).
				Return(sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRememberingASettlementBeforeEach(t)
			initSuite(suite)

			err := suite.sut.RememberTransaction(context.Background(), payment, txHash)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given an authorization expiring before the lock ttl", func(t *testing.T) {
		t.Parallel()

		payment := paymentExpiringIn(time.Minute)

		initSuite := func(suite *WhenRememberingASettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().
				Store(gomock.Any(), txKey(payment), stored(payment), 10*time.Minute).
				Return(nil)
		}

		t.Run("Should remember the transaction for the lock ttl", func(t *testing.T) {
			t.Parallel()

			suite := WhenRememberingASettlementBeforeEach(t)
			initSuite(suite)

			err := suite.sut.RememberTransaction(context.Background(), payment, txHash)

			assert.NoError(t, err)
		})
	})

	t.Run("Given an authorization valid for days", func(t *testing.T) {
		t.Parallel()

		payment := paymentExpiringIn(48 * time.Hour)

		initSuite := func(suite *WhenRememberingASettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().Store(gomock.Any(), txKey(payment), stored(payment), time.Hour).Return(nil)
		}

		t.Run("Should remember the transaction for an hour", func(t *testing.T) {
			t.Parallel()

			suite := WhenRememberingASettlementBeforeEach(t)
			initSuite(suite)

			err := suite.sut.RememberTransaction(context.Background(), payment, txHash)

			assert.NoError(t, err)
		})
	})

	t.Run("Given an authorization expiring within the hour", func(t *testing.T) {
		t.Parallel()

		payment := paymentExpiringIn(30 * time.Minute)

		initSuite := func(suite *WhenRememberingASettlementTestingSuite) {
			suite.mockCacheStorage.EXPECT().
				Store(gomock.Any(), txKey(payment), stored(payment), gomock.Cond(func(ttl time.Duration) bool {
					return ttl > 29*time.Minute && ttl <= 30*time.Minute
				})).
				Return(nil)
		}

		t.Run("Should remember the transaction until it expires", func(t *testing.T) {
			t.Parallel()

			suite := WhenRememberingASettlementBeforeEach(t)
			initSuite(suite)

			err := suite.sut.RememberTransaction(context.Background(), payment, txHash)

			assert.NoError(t, err)
		})
	})
}
