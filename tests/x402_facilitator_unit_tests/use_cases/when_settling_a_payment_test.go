package use_cases_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"go.uber.org/zap"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/responses"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/metrics_mocks/interfaces_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenSettlingAPaymentTestingSuite struct {
	mockPaymentVerifier    *interfaces_mocks.MockPaymentVerifierInterface
	mockSettlementGuard    *interfaces_mocks.MockSettlementGuardInterface
	mockTransferSubmitter  *interfaces_mocks.MockTransferSubmitterInterface
	mockFacilitatorMetrics *metrics_mocks.MockFacilitatorMetricsInterface
	sut                    *use_cases.SettlePayment
}

func WhenSettlingAPaymentBeforeEach(t *testing.T) *WhenSettlingAPaymentTestingSuite {
	mockController := gomock.NewController(t)
	mockPaymentVerifier := interfaces_mocks.NewMockPaymentVerifierInterface(mockController)
	mockSettlementGuard := interfaces_mocks.NewMockSettlementGuardInterface(mockController)
	mockTransferSubmitter := interfaces_mocks.NewMockTransferSubmitterInterface(mockController)
	strategies := x402_facilitator_services.NewAssetTransferMethodStrategies(
		x402_facilitator_services.NewAssetTransferMethodStrategiesParams{
			Strategies: []x402_facilitator_interfaces.AssetTransferMethodStrategyInterface{
				x402_facilitator_services.NewEip3009AssetTransferMethodStrategy(zap.NewNop(), nil),
			},
		},
	)

	mockFacilitatorMetrics := metrics_mocks.NewMockFacilitatorMetricsInterface(mockController)

	return &WhenSettlingAPaymentTestingSuite{
		mockPaymentVerifier:    mockPaymentVerifier,
		mockSettlementGuard:    mockSettlementGuard,
		mockTransferSubmitter:  mockTransferSubmitter,
		mockFacilitatorMetrics: mockFacilitatorMetrics,
		sut: use_cases.NewSettlePayment(
			strategies,
			mockPaymentVerifier,
			mockSettlementGuard,
			mockTransferSubmitter,
			mockFacilitatorMetrics,
		),
	}
}

func (suite *WhenSettlingAPaymentTestingSuite) paymentKeyed(key string) gomock.Matcher {
	return gomock.Cond(func(payment *models.Payment) bool { return payment.SettlementKey() == key })
}

func (suite *WhenSettlingAPaymentTestingSuite) expectAuthorized() {
	suite.mockPaymentVerifier.EXPECT().VerifyAuthorization(gomock.Any()).Return(nil)
}

func (suite *WhenSettlingAPaymentTestingSuite) expectNothingRemembered(key string) {
	suite.mockSettlementGuard.EXPECT().RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).Return(nil, nil)
}

func (suite *WhenSettlingAPaymentTestingSuite) expectLocked(key string) {
	suite.expectNothingRemembered(key)
	suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
	suite.mockSettlementGuard.EXPECT().TryAcquire(gomock.Any(), key).Return("token", true, nil)
	suite.expectNothingRemembered(key)
}

func (suite *WhenSettlingAPaymentTestingSuite) expectBroadcast(txHash common.Hash, err error) {
	suite.mockTransferSubmitter.EXPECT().
		Broadcast(gomock.Any(), uint64(80002), gomock.Any()).
		Return(txHash, err)
}

func (suite *WhenSettlingAPaymentTestingSuite) expectRememberedAndReleased(key string, txHash common.Hash) {
	suite.mockSettlementGuard.EXPECT().RememberTransaction(gomock.Any(), suite.paymentKeyed(key), txHash).Return(nil)
	suite.mockSettlementGuard.EXPECT().Release(gomock.Any(), key, "token")
}

func (suite *WhenSettlingAPaymentTestingSuite) expectAwaited(txHash common.Hash, status models.TransferStatus) {
	suite.mockTransferSubmitter.EXPECT().
		AwaitTransfer(gomock.Any(), uint64(80002), txHash).
		Return(&models.Transfer{TxHash: txHash, Status: status}, nil)
}

func (suite *WhenSettlingAPaymentTestingSuite) expectStatus(txHash common.Hash, status models.TransferStatus) {
	suite.mockTransferSubmitter.EXPECT().
		TransferStatus(gomock.Any(), uint64(80002), txHash).
		Return(&models.Transfer{TxHash: txHash, Status: status}, nil)
}

func (suite *WhenSettlingAPaymentTestingSuite) expectSettlement(verified bool, outcome string) {
	suite.mockFacilitatorMetrics.EXPECT().
		RecordSettlement(gomock.Any(), gomock.Not(gomock.Nil()), verified, outcome, gomock.Any(), nil)
}

func (suite *WhenSettlingAPaymentTestingSuite) expectFailure(verified bool) {
	suite.mockFacilitatorMetrics.EXPECT().
		RecordSettlement(gomock.Any(), gomock.Not(gomock.Nil()), verified, "", gomock.Any(), gomock.Any())
}

func (suite *WhenSettlingAPaymentTestingSuite) assertUnavailable(t *testing.T, err error, code string) {
	var unavailableErr *apperrors.UnavailableError
	require.ErrorAs(t, err, &unavailableErr)
	assert.Equal(t, code, unavailableErr.Code)
}

func (suite *WhenSettlingAPaymentTestingSuite) assertRefused(
	t *testing.T,
	response *responses.SettlePaymentResponse,
	err error,
	reason string,
	transaction string,
) {
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, reason, response.ErrorReason)
	assert.Equal(t, transaction, response.Transaction)
}

func TestWhenSettlingAPayment(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0xabc")

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildSettleRequest()
		request.X402Version = 1

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.mockFacilitatorMetrics.EXPECT().
				RecordSettlement(
					gomock.Any(),
					gomock.Nil(),
					false,
					enums.PaymentReasonInvalidX402Version.String(),
					gomock.Any(),
					nil,
				)
		}

		t.Run("Should answer success false with the reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidX402Version.String(), "")
		})
	})

	t.Run("Given no strategy settles the requested transfer method", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().WithPermit2().BuildSettleRequest()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.mockFacilitatorMetrics.EXPECT().
				RecordSettlement(
					gomock.Any(),
					gomock.Nil(),
					false,
					enums.PaymentReasonInvalidPaymentRequirements.String(),
					gomock.Any(),
					nil,
				)
		}

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReasonInvalidPaymentRequirements.String(), response.ErrorReason)
		})
	})

	t.Run("Given a signature that does not match the authorization", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildSettleRequest()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.mockPaymentVerifier.EXPECT().VerifyAuthorization(gomock.Any()).
				Return(apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidSignature, "signature does not recover to from"))
			suite.expectSettlement(false, enums.PaymentReasonInvalidSignature.String())
		}

		t.Run("Should answer success false without reading the remembered transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidSignature.String(), "")
		})
	})

	t.Run("Given its nonce was already used by another authorization", func(t *testing.T) {
		t.Parallel()

		request := model_builders.NewPaymentRequestBuilder().BuildSettleRequest()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.mockSettlementGuard.EXPECT().RememberedTransaction(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidTransactionState, "nonce already used"))
			suite.expectSettlement(false, enums.PaymentReasonInvalidTransactionState.String())
		}

		t.Run("Should answer success false without a transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidTransactionState.String(), "")
		})
	})

	t.Run("Given the remembered transaction cannot be read", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(nil, errors.New("redis down"))
			suite.expectFailure(false)
		}

		t.Run("Should return store_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "store_unavailable")
		})
	})

	t.Run("Given the status of the transaction already sent cannot be read", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(&txHash, nil)
			suite.mockTransferSubmitter.EXPECT().
				TransferStatus(gomock.Any(), uint64(80002), txHash).
				Return(nil, errors.New("rpc down"))
			suite.expectFailure(false)
		}

		t.Run("Should return chain_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "chain_unavailable")
		})
	})

	t.Run("Given the transaction already sent is confirmed", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(&txHash, nil)
			suite.expectStatus(txHash, models.TransferStatusConfirmed)
			suite.expectSettlement(true, enums.PaymentReasonInvalidTransactionState.String())
		}

		t.Run("Should answer success false with the transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidTransactionState.String(), txHash.Hex())
		})
	})

	t.Run("Given the transaction already sent reverted", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(&txHash, nil)
			suite.expectStatus(txHash, models.TransferStatusReverted)
			suite.expectSettlement(true, enums.PaymentReasonInvalidTransactionState.String())
		}

		t.Run("Should answer success false with the transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidTransactionState.String(), txHash.Hex())
		})
	})

	t.Run("Given the transaction already sent is still pending", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(&txHash, nil)
			suite.expectStatus(txHash, models.TransferStatusPending)
			suite.expectSettlement(true, enums.PaymentReasonSettlementPending.String())
		}

		t.Run("Should answer settlement_pending with the transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonSettlementPending.String(), txHash.Hex())
		})
	})

	t.Run("Given the verifier refuses the payment", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectNothingRemembered(key)
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).
				Return(apperrors.NewInvalidPaymentError(enums.PaymentReasonInsufficientFunds, "payer balance is below the amount"))
			suite.expectSettlement(false, enums.PaymentReasonInsufficientFunds.String())
		}

		t.Run("Should answer success false with the reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInsufficientFunds.String(), "")
		})
	})

	t.Run("Given the verifier cannot reach a dependency", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()
		unavailableErr := apperrors.NewUnavailableError(
			"registry_unavailable",
			"the agent registry is unavailable",
			nil,
		)

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectNothingRemembered(key)
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(unavailableErr)
			suite.mockFacilitatorMetrics.EXPECT().
				RecordSettlement(gomock.Any(), gomock.Not(gomock.Nil()), false, "", gomock.Any(), unavailableErr)
		}

		t.Run("Should return the unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			assert.Same(t, unavailableErr, err)
		})
	})

	t.Run("Given the settlement lock cannot be taken", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectNothingRemembered(key)
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().TryAcquire(gomock.Any(), key).Return("", false, errors.New("redis down"))
			suite.expectFailure(true)
		}

		t.Run("Should return store_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "store_unavailable")
		})
	})

	t.Run("Given a settlement of the same authorization is in progress", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectNothingRemembered(key)
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().TryAcquire(gomock.Any(), key).Return("", false, nil)
			suite.expectFailure(true)
		}

		t.Run("Should return settlement_in_progress", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "settlement_in_progress")
		})
	})

	t.Run("Given another replica sent other terms under the same nonce before the lock was taken", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectNothingRemembered(key)
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().TryAcquire(gomock.Any(), key).Return("token", true, nil)
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(nil, apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidTransactionState, "nonce already used"))
			suite.mockSettlementGuard.EXPECT().Release(gomock.Any(), key, "token")
			suite.expectSettlement(true, enums.PaymentReasonInvalidTransactionState.String())
		}

		t.Run("Should answer success false without sending", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidTransactionState.String(), "")
		})
	})

	t.Run("Given the remembered transaction cannot be read again under the lock", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectNothingRemembered(key)
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().TryAcquire(gomock.Any(), key).Return("token", true, nil)
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(nil, errors.New("redis down"))
			suite.mockSettlementGuard.EXPECT().Release(gomock.Any(), key, "token")
			suite.expectFailure(true)
		}

		t.Run("Should return store_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "store_unavailable")
		})
	})

	t.Run("Given another replica sent it before the lock was taken", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectNothingRemembered(key)
			suite.mockPaymentVerifier.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
			suite.mockSettlementGuard.EXPECT().TryAcquire(gomock.Any(), key).Return("token", true, nil)
			suite.mockSettlementGuard.EXPECT().
				RememberedTransaction(gomock.Any(), suite.paymentKeyed(key)).
				Return(&txHash, nil)
			suite.mockSettlementGuard.EXPECT().Release(gomock.Any(), key, "token")
			suite.expectStatus(txHash, models.TransferStatusPending)
			suite.expectSettlement(true, enums.PaymentReasonSettlementPending.String())
		}

		t.Run("Should report that transaction without sending again", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonSettlementPending.String(), txHash.Hex())
		})
	})

	t.Run("Given the chain rejects the transfer", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(common.Hash{}, fmt.Errorf("%w: execution reverted", apperrors.ErrTransferRejected))
			suite.mockSettlementGuard.EXPECT().Release(gomock.Any(), key, "token")
			suite.expectSettlement(true, enums.PaymentReasonInvalidTransactionState.String())
		}

		t.Run("Should answer success false with the reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidTransactionState.String(), "")
		})
	})

	t.Run("Given the relayer wallet cannot be loaded", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(common.Hash{}, fmt.Errorf("%w: vault sealed", apperrors.ErrRelayerUnavailable))
			suite.mockSettlementGuard.EXPECT().Release(gomock.Any(), key, "token")
			suite.expectFailure(true)
		}

		t.Run("Should return relayer_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "relayer_unavailable")
		})
	})

	t.Run("Given the transfer cannot be sent", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(common.Hash{}, errors.New("rpc down"))
			suite.mockSettlementGuard.EXPECT().Release(gomock.Any(), key, "token")
			suite.expectFailure(true)
		}

		t.Run("Should return settlement_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "settlement_unavailable")
		})
	})

	t.Run("Given the transaction sent cannot be remembered", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(txHash, nil)
			suite.mockSettlementGuard.EXPECT().RememberTransaction(gomock.Any(), suite.paymentKeyed(key), txHash).
				Return(errors.New("redis down"))
			suite.expectAwaited(txHash, models.TransferStatusConfirmed)
			suite.expectSettlement(true, "settled")
		}

		t.Run("Should not release the lock", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			require.NoError(t, err)
		})
	})

	t.Run("Given the outcome cannot be awaited", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(txHash, nil)
			suite.expectRememberedAndReleased(key, txHash)
			suite.mockTransferSubmitter.EXPECT().
				AwaitTransfer(gomock.Any(), uint64(80002), txHash).
				Return(nil, errors.New("no RPC"))
			suite.expectFailure(true)
		}

		t.Run("Should return chain_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "chain_unavailable")
		})
	})

	t.Run("Given the transfer reverts on-chain", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(txHash, nil)
			suite.expectRememberedAndReleased(key, txHash)
			suite.expectAwaited(txHash, models.TransferStatusReverted)
			suite.expectSettlement(true, enums.PaymentReasonInvalidTransactionState.String())
		}

		t.Run("Should answer success false with the transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonInvalidTransactionState.String(), txHash.Hex())
		})
	})

	t.Run("Given the transfer is not confirmed in time", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(txHash, nil)
			suite.expectRememberedAndReleased(key, txHash)
			suite.expectAwaited(txHash, models.TransferStatusPending)
			suite.expectSettlement(true, enums.PaymentReasonSettlementPending.String())
		}

		t.Run("Should answer settlement_pending with the transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			suite.assertRefused(t, response, err, enums.PaymentReasonSettlementPending.String(), txHash.Hex())
		})
	})

	t.Run("Given the transfer is confirmed", func(t *testing.T) {
		t.Parallel()

		builder := model_builders.NewPaymentRequestBuilder()
		request := builder.BuildSettleRequest()
		key := builder.BuildPayment().SettlementKey()

		initSuite := func(suite *WhenSettlingAPaymentTestingSuite) {
			suite.expectAuthorized()
			suite.expectLocked(key)
			suite.expectBroadcast(txHash, nil)
			suite.expectRememberedAndReleased(key, txHash)
			suite.expectAwaited(txHash, models.TransferStatusConfirmed)
			suite.expectSettlement(true, "settled")
		}

		t.Run("Should answer success with the transaction, payer and network", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingAPaymentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			require.NoError(t, err)
			assert.True(t, response.Success)
			assert.Equal(t, txHash.Hex(), response.Transaction)
			assert.Equal(t, builder.Payer().Hex(), response.Payer)
			assert.Equal(t, "eip155:80002", response.Network)
		})
	})
}
