package services_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenReservingGasTestingSuite struct {
	gasBudgetSettings  *x402_facilitator_settings.GasBudgetSettings
	mockCounterStorage *interfaces_mocks.MockCounterStorageInterface
	mockCacheStorage   *interfaces_mocks.MockCacheStorageInterface
	logs               *observer.ObservedLogs
	sut                *x402_facilitator_services.GasBudget
}

func WhenReservingGasBeforeEach(t *testing.T) *WhenReservingGasTestingSuite {
	mockController := gomock.NewController(t)
	gasBudgetSettings := &x402_facilitator_settings.GasBudgetSettings{80002: 1000}
	mockCounterStorage := interfaces_mocks.NewMockCounterStorageInterface(mockController)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)
	core, logs := observer.New(zapcore.WarnLevel)

	return &WhenReservingGasTestingSuite{
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

func (suite *WhenReservingGasTestingSuite) dayKey() gomock.Matcher {
	return gomock.Cond(func(key string) bool {
		return key == fmt.Sprintf("x402:gas:80002:%d", time.Now().Unix()/86400)
	})
}

func (suite *WhenReservingGasTestingSuite) reservation(amount int64) gomock.Matcher {
	return gomock.Cond(func(value string) bool {
		return value == fmt.Sprintf("x402:gas:80002:%d %d", time.Now().Unix()/86400, amount)
	})
}

func TestWhenReservingGas(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0xabc")

	t.Run("Given the chain has no gas budget", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept the cost without touching the store", func(t *testing.T) {
			t.Parallel()

			suite := WhenReservingGasBeforeEach(t)

			err := suite.sut.Reserve(context.Background(), 137, txHash, big.NewInt(600_000_000_000))

			assert.NoError(t, err)
		})
	})

	t.Run("Given a cost above the whole daily budget", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the budget spent error and count nothing", func(t *testing.T) {
			t.Parallel()

			suite := WhenReservingGasBeforeEach(t)

			err := suite.sut.Reserve(context.Background(), 80002, txHash, new(big.Int).Lsh(big.NewInt(1), 200))

			assert.ErrorIs(t, err, apperrors.ErrGasBudgetSpent)
		})
	})

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("redis down")

		initSuite := func(suite *WhenReservingGasTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.dayKey(), int64(600), int64(1000), 24*time.Hour).
				Return(false, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReservingGasBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reserve(context.Background(), 80002, txHash, big.NewInt(600_000_000_000))

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given a cost above what is left of the budget", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReservingGasTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.dayKey(), int64(600), int64(1000), 24*time.Hour).
				Return(false, nil)
		}

		t.Run("Should return the budget spent error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReservingGasBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reserve(context.Background(), 80002, txHash, big.NewInt(600_000_000_000))

			assert.ErrorIs(t, err, apperrors.ErrGasBudgetSpent)
		})
	})

	t.Run("Given the cost is counted but its reservation cannot be stored", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReservingGasTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.dayKey(), int64(600), int64(1000), 24*time.Hour).
				Return(true, nil)
			suite.mockCacheStorage.EXPECT().
				Store(
					gomock.Any(),
					"x402:gas:80002:tx:"+txHash.Hex(),
					suite.reservation(600),
					24*time.Hour,
				).
				Return(errors.New("redis down"))
		}

		t.Run("Should accept the cost and log the failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenReservingGasBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reserve(context.Background(), 80002, txHash, big.NewInt(600_000_000_000))

			assert.NoError(t, err)
			assert.Equal(t, 1, suite.logs.Len())
		})
	})

	t.Run("Given a cost that is not a whole number of gwei", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReservingGasTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.dayKey(), int64(601), int64(1000), 24*time.Hour).
				Return(true, nil)
			suite.mockCacheStorage.EXPECT().
				Store(
					gomock.Any(),
					"x402:gas:80002:tx:"+txHash.Hex(),
					suite.reservation(601),
					24*time.Hour,
				).
				Return(nil)
		}

		t.Run("Should count it rounded up to the next gwei", func(t *testing.T) {
			t.Parallel()

			suite := WhenReservingGasBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reserve(context.Background(), 80002, txHash, big.NewInt(600_000_000_001))

			assert.NoError(t, err)
		})
	})

	t.Run("Given a cost within what is left of the budget", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReservingGasTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.dayKey(), int64(600), int64(1000), 24*time.Hour).
				Return(true, nil)
			suite.mockCacheStorage.EXPECT().
				Store(
					gomock.Any(),
					"x402:gas:80002:tx:"+txHash.Hex(),
					suite.reservation(600),
					24*time.Hour,
				).
				Return(nil)
		}

		t.Run("Should count the cost and store its reservation", func(t *testing.T) {
			t.Parallel()

			suite := WhenReservingGasBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reserve(context.Background(), 80002, txHash, big.NewInt(600_000_000_000))

			assert.NoError(t, err)
			assert.Equal(t, 0, suite.logs.Len())
		})
	})
}
