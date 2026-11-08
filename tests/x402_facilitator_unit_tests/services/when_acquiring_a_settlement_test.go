package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenAcquiringASettlementTestingSuite struct {
	mockLockStorage  *interfaces_mocks.MockLockStorageInterface
	mockCacheStorage *interfaces_mocks.MockCacheStorageInterface
	sut              *x402_facilitator_services.SettlementGuard
}

func WhenAcquiringASettlementBeforeEach(t *testing.T) *WhenAcquiringASettlementTestingSuite {
	mockController := gomock.NewController(t)
	mockLockStorage := interfaces_mocks.NewMockLockStorageInterface(mockController)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)

	return &WhenAcquiringASettlementTestingSuite{
		mockLockStorage:  mockLockStorage,
		mockCacheStorage: mockCacheStorage,
		sut: x402_facilitator_services.NewSettlementGuard(
			mockLockStorage,
			mockCacheStorage,
			&x402_facilitator_settings.SettlementSettings{LockTtl: 10 * time.Minute},
			zap.NewNop(),
		),
	}
}

func TestWhenAcquiringASettlement(t *testing.T) {
	t.Parallel()

	const key = "80002:token:payer:nonce"
	const lockKey = "x402:settlement:" + key

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("redis down")

		initSuite := func(suite *WhenAcquiringASettlementTestingSuite) {
			suite.mockLockStorage.EXPECT().TryLock(gomock.Any(), lockKey, 10*time.Minute).Return("", false, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAcquiringASettlementBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.sut.TryAcquire(context.Background(), key)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given another settlement holds the key", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAcquiringASettlementTestingSuite) {
			suite.mockLockStorage.EXPECT().TryLock(gomock.Any(), lockKey, 10*time.Minute).Return("", false, nil)
		}

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenAcquiringASettlementBeforeEach(t)
			initSuite(suite)

			_, acquired, err := suite.sut.TryAcquire(context.Background(), key)

			assert.NoError(t, err)
			assert.False(t, acquired)
		})
	})

	t.Run("Given the key is free", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAcquiringASettlementTestingSuite) {
			suite.mockLockStorage.EXPECT().TryLock(gomock.Any(), lockKey, 10*time.Minute).Return("token", true, nil)
		}

		t.Run("Should return the lock token", func(t *testing.T) {
			t.Parallel()

			suite := WhenAcquiringASettlementBeforeEach(t)
			initSuite(suite)

			token, acquired, err := suite.sut.TryAcquire(context.Background(), key)

			assert.NoError(t, err)
			assert.True(t, acquired)
			assert.Equal(t, "token", token)
		})
	})
}
