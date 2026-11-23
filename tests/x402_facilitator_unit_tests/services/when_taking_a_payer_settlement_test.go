package services_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenTakingAPayerSettlementTestingSuite struct {
	settlementSettings *x402_facilitator_settings.SettlementSettings
	mockCounterStorage *interfaces_mocks.MockCounterStorageInterface
	sut                *x402_facilitator_services.PayerLimiter
}

func WhenTakingAPayerSettlementBeforeEach(t *testing.T) *WhenTakingAPayerSettlementTestingSuite {
	settlementSettings := &x402_facilitator_settings.SettlementSettings{MaxPerPayerPerHour: 3}
	mockCounterStorage := interfaces_mocks.NewMockCounterStorageInterface(gomock.NewController(t))

	return &WhenTakingAPayerSettlementTestingSuite{
		settlementSettings: settlementSettings,
		mockCounterStorage: mockCounterStorage,
		sut: x402_facilitator_services.NewPayerLimiter(
			settlementSettings,
			mockCounterStorage,
			zap.NewNop(),
		),
	}
}

func (suite *WhenTakingAPayerSettlementTestingSuite) keyOf(payment *models.Payment) gomock.Matcher {
	payer := strings.ToLower(payment.Authorization.From.Hex())
	return gomock.Cond(func(key string) bool {
		return key == fmt.Sprintf("x402:payer:80002:%s:%d", payer, time.Now().Unix()/3600)
	})
}

func TestWhenTakingAPayerSettlement(t *testing.T) {
	t.Parallel()

	t.Run("Given no payer limit is set", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		initSuite := func(suite *WhenTakingAPayerSettlementTestingSuite) {
			suite.settlementSettings.MaxPerPayerPerHour = 0
		}

		t.Run("Should return no hold and count nothing", func(t *testing.T) {
			t.Parallel()

			suite := WhenTakingAPayerSettlementBeforeEach(t)
			initSuite(suite)

			hold, err := suite.sut.Take(context.Background(), payment)

			assert.NoError(t, err)
			assert.Empty(t, hold)
		})
	})

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()
		sourceErr := errors.New("redis down")

		initSuite := func(suite *WhenTakingAPayerSettlementTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.keyOf(payment), int64(1), int64(3), time.Hour).
				Return(false, sourceErr)
		}

		t.Run("Should return store_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenTakingAPayerSettlementBeforeEach(t)
			initSuite(suite)

			hold, err := suite.sut.Take(context.Background(), payment)

			assert.Empty(t, hold)
			var unavailableErr *apperrors.UnavailableError
			require.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "store_unavailable", unavailableErr.Code)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the payer reached its limit", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		initSuite := func(suite *WhenTakingAPayerSettlementTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.keyOf(payment), int64(1), int64(3), time.Hour).
				Return(false, nil)
		}

		t.Run("Should refuse the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenTakingAPayerSettlementBeforeEach(t)
			initSuite(suite)

			hold, err := suite.sut.Take(context.Background(), payment)

			assert.Empty(t, hold)
			var invalidPayment *apperrors.InvalidPaymentError
			require.ErrorAs(t, err, &invalidPayment)
			assert.Equal(t, enums.PaymentReasonPayerLimitReached.String(), invalidPayment.Reason)
		})
	})

	t.Run("Given the payer is below its limit", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		initSuite := func(suite *WhenTakingAPayerSettlementTestingSuite) {
			suite.mockCounterStorage.EXPECT().
				TryAdd(gomock.Any(), suite.keyOf(payment), int64(1), int64(3), time.Hour).
				Return(true, nil)
		}

		t.Run("Should count the settlement and return the key it was counted under", func(t *testing.T) {
			t.Parallel()

			suite := WhenTakingAPayerSettlementBeforeEach(t)
			initSuite(suite)

			hold, err := suite.sut.Take(context.Background(), payment)

			assert.NoError(t, err)
			assert.True(t, suite.keyOf(payment).Matches(hold))
		})
	})
}
