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

type WhenCheckingAPayerLimitTestingSuite struct {
	settlementSettings *x402_facilitator_settings.SettlementSettings
	mockCounterStorage *interfaces_mocks.MockCounterStorageInterface
	sut                *x402_facilitator_services.PayerLimiter
}

func WhenCheckingAPayerLimitBeforeEach(t *testing.T) *WhenCheckingAPayerLimitTestingSuite {
	settlementSettings := &x402_facilitator_settings.SettlementSettings{MaxPerPayerPerHour: 3}
	mockCounterStorage := interfaces_mocks.NewMockCounterStorageInterface(gomock.NewController(t))

	return &WhenCheckingAPayerLimitTestingSuite{
		settlementSettings: settlementSettings,
		mockCounterStorage: mockCounterStorage,
		sut: x402_facilitator_services.NewPayerLimiter(
			settlementSettings,
			mockCounterStorage,
			zap.NewNop(),
		),
	}
}

func (suite *WhenCheckingAPayerLimitTestingSuite) keyOf(payment *models.Payment) gomock.Matcher {
	payer := strings.ToLower(payment.Authorization.From.Hex())
	return gomock.Cond(func(key string) bool {
		return key == fmt.Sprintf("x402:payer:80002:%s:%d", payer, time.Now().Unix()/3600)
	})
}

func TestWhenCheckingAPayerLimit(t *testing.T) {
	t.Parallel()

	t.Run("Given no payer limit is set", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		initSuite := func(suite *WhenCheckingAPayerLimitTestingSuite) {
			suite.settlementSettings.MaxPerPayerPerHour = 0
		}

		t.Run("Should accept the payment without reading the store", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAPayerLimitBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payment)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the store cannot be reached", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()
		sourceErr := errors.New("redis down")

		initSuite := func(suite *WhenCheckingAPayerLimitTestingSuite) {
			suite.mockCounterStorage.EXPECT().Read(gomock.Any(), suite.keyOf(payment)).Return(int64(0), sourceErr)
		}

		t.Run("Should return store_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAPayerLimitBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payment)

			var unavailableErr *apperrors.UnavailableError
			require.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "store_unavailable", unavailableErr.Code)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the payer reached its limit", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		initSuite := func(suite *WhenCheckingAPayerLimitTestingSuite) {
			suite.mockCounterStorage.EXPECT().Read(gomock.Any(), suite.keyOf(payment)).Return(int64(3), nil)
		}

		t.Run("Should refuse the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAPayerLimitBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payment)

			var invalidPayment *apperrors.InvalidPaymentError
			require.ErrorAs(t, err, &invalidPayment)
			assert.Equal(t, enums.PaymentReasonPayerLimitReached.String(), invalidPayment.Reason)
		})
	})

	t.Run("Given the payer is below its limit", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		initSuite := func(suite *WhenCheckingAPayerLimitTestingSuite) {
			suite.mockCounterStorage.EXPECT().Read(gomock.Any(), suite.keyOf(payment)).Return(int64(2), nil)
		}

		t.Run("Should accept the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAPayerLimitBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payment)

			assert.NoError(t, err)
		})
	})
}
