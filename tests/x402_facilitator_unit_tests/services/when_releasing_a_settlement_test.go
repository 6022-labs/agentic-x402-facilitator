package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReleasingASettlementTestingSuite struct {
	mockLockStorage  *interfaces_mocks.MockLockStorageInterface
	mockCacheStorage *interfaces_mocks.MockCacheStorageInterface
	sut              *x402_facilitator_services.SettlementGuard
}

func WhenReleasingASettlementBeforeEach(t *testing.T) *WhenReleasingASettlementTestingSuite {
	mockController := gomock.NewController(t)
	mockLockStorage := interfaces_mocks.NewMockLockStorageInterface(mockController)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)

	return &WhenReleasingASettlementTestingSuite{
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

func TestWhenReleasingASettlement(t *testing.T) {
	t.Parallel()

	const key = "80002:token:payer:nonce"
	const lockKey = "x402:settlement:" + key

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		t.Run("Should swallow the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingASettlementBeforeEach(t)
			suite.mockLockStorage.EXPECT().Unlock(gomock.Any(), lockKey, "token").Return(errors.New("redis down"))

			suite.sut.Release(context.Background(), key, "token")
		})
	})

	t.Run("Given the settlement holds the key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should unlock it with its token", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingASettlementBeforeEach(t)
			suite.mockLockStorage.EXPECT().Unlock(gomock.Any(), lockKey, "token").Return(nil)

			suite.sut.Release(context.Background(), key, "token")
		})
	})
}
