package services_test

import (
	"context"
	"errors"
	"testing"

	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenGivingBackAPayerSettlementTestingSuite struct {
	mockCounterStorage *interfaces_mocks.MockCounterStorageInterface
	logs               *observer.ObservedLogs
	sut                *x402_facilitator_services.PayerLimiter
}

func WhenGivingBackAPayerSettlementBeforeEach(t *testing.T) *WhenGivingBackAPayerSettlementTestingSuite {
	mockCounterStorage := interfaces_mocks.NewMockCounterStorageInterface(gomock.NewController(t))
	core, logs := observer.New(zapcore.WarnLevel)

	return &WhenGivingBackAPayerSettlementTestingSuite{
		mockCounterStorage: mockCounterStorage,
		logs:               logs,
		sut: x402_facilitator_services.NewPayerLimiter(
			&x402_facilitator_settings.SettlementSettings{MaxPerPayerPerHour: 3},
			mockCounterStorage,
			zap.New(core),
		),
	}
}

func TestWhenGivingBackAPayerSettlement(t *testing.T) {
	t.Parallel()

	t.Run("Given no settlement was counted", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not touch the store", func(t *testing.T) {
			t.Parallel()

			suite := WhenGivingBackAPayerSettlementBeforeEach(t)

			suite.sut.GiveBack(context.Background(), "")

			assert.Equal(t, 0, suite.logs.Len())
		})
	})

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGivingBackAPayerSettlementTestingSuite) {
			suite.mockCounterStorage.EXPECT().Subtract(gomock.Any(), "hold", int64(1)).Return(errors.New("redis down"))
		}

		t.Run("Should log the failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenGivingBackAPayerSettlementBeforeEach(t)
			initSuite(suite)

			suite.sut.GiveBack(context.Background(), "hold")

			assert.Equal(t, 1, suite.logs.Len())
		})
	})

	t.Run("Given a counted settlement", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGivingBackAPayerSettlementTestingSuite) {
			suite.mockCounterStorage.EXPECT().Subtract(gomock.Any(), "hold", int64(1)).Return(nil)
		}

		t.Run("Should take it off the payer's count", func(t *testing.T) {
			t.Parallel()

			suite := WhenGivingBackAPayerSettlementBeforeEach(t)
			initSuite(suite)

			suite.sut.GiveBack(context.Background(), "hold")

			assert.Equal(t, 0, suite.logs.Len())
		})
	})

	t.Run("Given a counted settlement and a cancelled request", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		initSuite := func(suite *WhenGivingBackAPayerSettlementTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				Subtract(gomock.Cond(func(ctx context.Context) bool { return ctx.Err() == nil }), "hold", int64(1)).
				Return(nil)
		}

		t.Run("Should still take it off the payer's count", func(t *testing.T) {
			t.Parallel()

			suite := WhenGivingBackAPayerSettlementBeforeEach(t)
			initSuite(suite)

			suite.sut.GiveBack(ctx, "hold")

			assert.Equal(t, 0, suite.logs.Len())
		})
	})
}
