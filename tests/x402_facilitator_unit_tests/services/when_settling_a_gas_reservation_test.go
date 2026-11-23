package services_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

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

type WhenSettlingAGasReservationTestingSuite struct {
	gasBudgetSettings  *x402_facilitator_settings.GasBudgetSettings
	mockCounterStorage *interfaces_mocks.MockCounterStorageInterface
	mockCacheStorage   *interfaces_mocks.MockCacheStorageInterface
	logs               *observer.ObservedLogs
	sut                *x402_facilitator_services.GasBudget
}

func WhenSettlingAGasReservationBeforeEach(t *testing.T) *WhenSettlingAGasReservationTestingSuite {
	mockController := gomock.NewController(t)
	gasBudgetSettings := &x402_facilitator_settings.GasBudgetSettings{80002: 1000}
	mockCounterStorage := interfaces_mocks.NewMockCounterStorageInterface(mockController)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)
	core, logs := observer.New(zapcore.WarnLevel)

	return &WhenSettlingAGasReservationTestingSuite{
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

func TestWhenSettlingAGasReservation(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0xabc")

	t.Run("Given the chain has no gas budget", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not touch the store", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)

			suite.sut.Settle(context.Background(), 137, txHash, big.NewInt(200_000_000_000))

			assert.Equal(t, 0, suite.logs.Len())
		})
	})

	t.Run("Given the reservation cannot be read", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettlingAGasReservationTestingSuite) {
			suite.mockCacheStorage.EXPECT().
				ReadAndDelete(gomock.Any(), "x402:gas:80002:tx:"+txHash.Hex()).
				Return(nil, errors.New("redis down"))
		}

		t.Run("Should log the failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)
			initSuite(suite)

			suite.sut.Settle(context.Background(), 80002, txHash, big.NewInt(200_000_000_000))

			assert.Equal(t, 1, suite.logs.Len())
		})
	})

	t.Run("Given no reservation is stored", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettlingAGasReservationTestingSuite) {
			suite.mockCacheStorage.EXPECT().
				ReadAndDelete(gomock.Any(), "x402:gas:80002:tx:"+txHash.Hex()).
				Return(nil, nil)
		}

		t.Run("Should leave the day's counter unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)
			initSuite(suite)

			suite.sut.Settle(context.Background(), 80002, txHash, big.NewInt(200_000_000_000))

			assert.Equal(t, 0, suite.logs.Len())
		})
	})

	t.Run("Given the stored reservation is malformed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettlingAGasReservationTestingSuite) {
			reservation := "day"
			suite.mockCacheStorage.EXPECT().
				ReadAndDelete(gomock.Any(), "x402:gas:80002:tx:"+txHash.Hex()).
				Return(&reservation, nil)
		}

		t.Run("Should log the failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)
			initSuite(suite)

			suite.sut.Settle(context.Background(), 80002, txHash, big.NewInt(200_000_000_000))

			assert.Equal(t, 1, suite.logs.Len())
		})
	})

	t.Run("Given the day's counter cannot be updated", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettlingAGasReservationTestingSuite) {
			reservation := "day 600"
			suite.mockCacheStorage.EXPECT().
				ReadAndDelete(gomock.Any(), "x402:gas:80002:tx:"+txHash.Hex()).
				Return(&reservation, nil)
			suite.mockCounterStorage.EXPECT().Subtract(gomock.Any(), "day", int64(400)).Return(errors.New("redis down"))
		}

		t.Run("Should log the failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)
			initSuite(suite)

			suite.sut.Settle(context.Background(), 80002, txHash, big.NewInt(200_000_000_000))

			assert.Equal(t, 1, suite.logs.Len())
		})
	})

	t.Run("Given a transaction that cost all it reserved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettlingAGasReservationTestingSuite) {
			reservation := "day 600"
			suite.mockCacheStorage.EXPECT().
				ReadAndDelete(gomock.Any(), "x402:gas:80002:tx:"+txHash.Hex()).
				Return(&reservation, nil)
		}

		t.Run("Should leave the day's counter unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)
			initSuite(suite)

			suite.sut.Settle(context.Background(), 80002, txHash, big.NewInt(600_000_000_000))

			assert.Equal(t, 0, suite.logs.Len())
		})
	})

	t.Run("Given a transaction never broadcast", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettlingAGasReservationTestingSuite) {
			reservation := "day 600"
			suite.mockCacheStorage.EXPECT().
				ReadAndDelete(gomock.Any(), "x402:gas:80002:tx:"+txHash.Hex()).
				Return(&reservation, nil)
			suite.mockCounterStorage.EXPECT().Subtract(gomock.Any(), "day", int64(600)).Return(nil)
		}

		t.Run("Should take the whole reservation off the day's counter", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)
			initSuite(suite)

			suite.sut.Settle(context.Background(), 80002, txHash, new(big.Int))

			assert.Equal(t, 0, suite.logs.Len())
		})
	})

	t.Run("Given a transaction that cost less than it reserved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettlingAGasReservationTestingSuite) {
			reservation := "day 600"
			suite.mockCacheStorage.EXPECT().
				ReadAndDelete(gomock.Any(), "x402:gas:80002:tx:"+txHash.Hex()).
				Return(&reservation, nil)
			suite.mockCounterStorage.EXPECT().Subtract(gomock.Any(), "day", int64(400)).Return(nil)
		}

		t.Run("Should take the difference off the day's counter", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAGasReservationBeforeEach(t)
			initSuite(suite)

			suite.sut.Settle(context.Background(), 80002, txHash, big.NewInt(200_000_000_000))

			assert.Equal(t, 0, suite.logs.Len())
		})
	})
}
