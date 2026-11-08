package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/metrics_mocks/interfaces_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenVerifyingAPaymentTestingSuite struct {
	mockPaymentVerifier    *interfaces_mocks.MockPaymentVerifierInterface
	mockSettlementGuard    *interfaces_mocks.MockSettlementGuardInterface
	mockFacilitatorMetrics *metrics_mocks.MockFacilitatorMetricsInterface
	sut                    *use_cases.VerifyPayment
}

func WhenVerifyingAPaymentBeforeEach(t *testing.T) *WhenVerifyingAPaymentTestingSuite {
	mockController := gomock.NewController(t)
	mockPaymentVerifier := interfaces_mocks.NewMockPaymentVerifierInterface(mockController)
	mockSettlementGuard := interfaces_mocks.NewMockSettlementGuardInterface(mockController)
	mockFacilitatorMetrics := metrics_mocks.NewMockFacilitatorMetricsInterface(mockController)

	return &WhenVerifyingAPaymentTestingSuite{
		mockPaymentVerifier:    mockPaymentVerifier,
		mockSettlementGuard:    mockSettlementGuard,
		mockFacilitatorMetrics: mockFacilitatorMetrics,
		sut: use_cases.NewVerifyPayment(
			mockPaymentVerifier,
			mockSettlementGuard,
			mockFacilitatorMetrics,
		),
	}
}

func (suite *WhenVerifyingAPaymentTestingSuite) expectAuthorizedAndUnsent() {
	suite.mockPaymentVerifier.EXPECT().VerifyAuthorization(gomock.Any()).Return(nil)
	suite.mockSettlementGuard.EXPECT().RememberedTransaction(gomock.Any(), gomock.Any()).Return(nil, nil)
}

func TestWhenVerifyingAPayment(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildVerifyRequest()
		request.X402Version = 1

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.mockFacilitatorMetrics.EXPECT().
				RecordVerification(
					gomock.Any(),
					gomock.Nil(),
					false,
					enums.PaymentReasonInvalidX402Version.String(),
					gomock.Any(),
					nil,
				)
		}

		t.Run("Should answer isValid false with the reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.False(t, response.IsValid)
			assert.Equal(t, enums.PaymentReasonInvalidX402Version.String(), response.InvalidReason)
			assert.Equal(t, builder.Payer().Hex(), response.Payer)
		})
	})

	t.Run("Given a signature that does not match the authorization", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildVerifyRequest()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.mockPaymentVerifier.EXPECT().VerifyAuthorization(gomock.Any()).
				Return(apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidSignature, "signature does not recover to from"))
			suite.mockFacilitatorMetrics.EXPECT().
				RecordVerification(
					gomock.Any(),
					gomock.Not(gomock.Nil()),
					false,
					enums.PaymentReasonInvalidSignature.String(),
					gomock.Any(),
					nil,
				)
		}

		t.Run("Should answer isValid false without reading the remembered transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.False(t, response.IsValid)
			assert.Equal(t, enums.PaymentReasonInvalidSignature.String(), response.InvalidReason)
		})
	})

	t.Run("Given its nonce was already used by another authorization", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildVerifyRequest()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.mockPaymentVerifier.EXPECT().VerifyAuthorization(gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().RememberedTransaction(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidTransactionState, "nonce already used"))
			suite.mockFacilitatorMetrics.EXPECT().RecordVerification(
				gomock.Any(),
				gomock.Not(gomock.Nil()),
				false,
				enums.PaymentReasonInvalidTransactionState.String(),
				gomock.Any(),
				nil,
			)
		}

		t.Run("Should answer isValid false", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.False(t, response.IsValid)
			assert.Equal(t, enums.PaymentReasonInvalidTransactionState.String(), response.InvalidReason)
		})
	})

	t.Run("Given the remembered transaction cannot be read", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildVerifyRequest()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.mockPaymentVerifier.EXPECT().VerifyAuthorization(gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().RememberedTransaction(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("redis down"))
			suite.mockFacilitatorMetrics.EXPECT().
				RecordVerification(gomock.Any(), gomock.Not(gomock.Nil()), false, "", gomock.Any(), gomock.Any())
		}

		t.Run("Should return store_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var unavailableErr *apperrors.UnavailableError
			require.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "store_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the authorization was already sent for settlement", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildVerifyRequest()
		txHash := common.HexToHash("0xabc")

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.mockPaymentVerifier.EXPECT().VerifyAuthorization(gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), gomock.Any()).
				Return(&txHash, nil)
			suite.mockFacilitatorMetrics.EXPECT().
				RecordVerification(
					gomock.Any(),
					gomock.Not(gomock.Nil()),
					false,
					enums.PaymentReasonInvalidTransactionState.String(),
					gomock.Any(),
					nil,
				)
		}

		t.Run("Should answer isValid false", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.False(t, response.IsValid)
			assert.Equal(t, enums.PaymentReasonInvalidTransactionState.String(), response.InvalidReason)
		})
	})

	t.Run("Given the verifier refuses the payment", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildVerifyRequest()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectAuthorizedAndUnsent()
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).
				Return(apperrors.NewInvalidPaymentError(enums.PaymentReasonInsufficientFunds, "payer balance is below the amount"))
			suite.mockFacilitatorMetrics.EXPECT().
				RecordVerification(
					gomock.Any(),
					gomock.Not(gomock.Nil()),
					false,
					enums.PaymentReasonInsufficientFunds.String(),
					gomock.Any(),
					nil,
				)
		}

		t.Run("Should answer isValid false with the reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.False(t, response.IsValid)
			assert.Equal(t, enums.PaymentReasonInsufficientFunds.String(), response.InvalidReason)
		})
	})

	t.Run("Given the verifier cannot reach a dependency", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildVerifyRequest()
		unavailableErr := apperrors.NewUnavailableError(
			"registry_unavailable",
			"the agent registry is unavailable",
			nil,
		)

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectAuthorizedAndUnsent()
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(unavailableErr)
			suite.mockFacilitatorMetrics.EXPECT().
				RecordVerification(gomock.Any(), gomock.Not(gomock.Nil()), false, "", gomock.Any(), unavailableErr)
		}

		t.Run("Should return the unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			assert.Same(t, unavailableErr, err)
		})
	})

	t.Run("Given the verifier accepts the payment", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildVerifyRequest()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectAuthorizedAndUnsent()
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), builder.BuildPayment()).Return(nil)
			suite.mockFacilitatorMetrics.EXPECT().
				RecordVerification(gomock.Any(), gomock.Not(gomock.Nil()), true, "valid", gomock.Any(), nil)
		}

		t.Run("Should verify the payment the request carries", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})

		t.Run("Should answer isValid true with the payer", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.True(t, response.IsValid)
			assert.Empty(t, response.InvalidReason)
			assert.Equal(t, builder.Payer().Hex(), response.Payer)
		})
	})
}
