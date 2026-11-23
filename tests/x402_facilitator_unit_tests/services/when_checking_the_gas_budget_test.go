package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenCheckingTheGasBudgetTestingSuite struct {
	gasBudgetSettings  *x402_facilitator_settings.GasBudgetSettings
	mockCounterStorage *interfaces_mocks.MockCounterStorageInterface
	mockCacheStorage   *interfaces_mocks.MockCacheStorageInterface
	logs               *observer.ObservedLogs
	sut                *x402_facilitator_services.GasBudget
}

func WhenCheckingTheGasBudgetBeforeEach(t *testing.T) *WhenCheckingTheGasBudgetTestingSuite {
	mockController := gomock.NewController(t)
	gasBudgetSettings := &x402_facilitator_settings.GasBudgetSettings{80002: 1000}
	mockCounterStorage := interfaces_mocks.NewMockCounterStorageInterface(mockController)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)
	core, logs := observer.New(zapcore.WarnLevel)

	return &WhenCheckingTheGasBudgetTestingSuite{
		gasBudgetSettings:  gasBudgetSettings,
		mockCounterStorage: mockCounterStorage,
		mockCacheStorage:   mockCacheStorage,
		logs:               logs,
		sut: x402_facilitator_services.NewGasBudget(
			gasBudgetSettings,
			mockCounterStorage,
			mockCacheStorage,
			zap.New(core),
		),
	}
}

func (suite *WhenCheckingTheGasBudgetTestingSuite) dayKey() gomock.Matcher {
	return gomock.Cond(func(key string) bool {
		return key == fmt.Sprintf("x402:gas:80002:%d", time.Now().Unix()/86400)
	})
}

func TestWhenCheckingTheGasBudget(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no gas budget", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept the settlement without reading the store", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingTheGasBudgetBeforeEach(t)

			err := suite.sut.Check(context.Background(), 137)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("redis down")

		initSuite := func(suite *WhenCheckingTheGasBudgetTestingSuite) {
			suite.mockCounterStorage.EXPECT().Read(gomock.Any(), suite.dayKey()).Return(int64(0), sourceErr)
		}

		t.Run("Should return store_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingTheGasBudgetBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), 80002)

			var unavailableErr *apperrors.UnavailableError
			require.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "store_unavailable", unavailableErr.Code)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the gas budget is spent", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingTheGasBudgetTestingSuite) {
			suite.mockCounterStorage.EXPECT().Read(gomock.Any(), suite.dayKey()).Return(int64(1000), nil)
		}

		t.Run("Should return gas_budget_spent", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingTheGasBudgetBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), 80002)

			var unavailableErr *apperrors.UnavailableError
			require.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "gas_budget_spent", unavailableErr.Code)
			assert.ErrorIs(t, err, apperrors.ErrGasBudgetSpent)
		})
	})

	t.Run("Given part of the gas budget is left", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingTheGasBudgetTestingSuite) {
			suite.mockCounterStorage.EXPECT().Read(gomock.Any(), suite.dayKey()).Return(int64(999), nil)
		}

		t.Run("Should accept the settlement", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingTheGasBudgetBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), 80002)

			assert.NoError(t, err)
		})
	})
}
