package services_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenVerifyingAPaymentTestingSuite struct {
	mockPayToEligibilityChecker *interfaces_mocks.MockPayToEligibilityCheckerInterface
	mockEip3009TokenRequester   *interfaces_mocks.MockEip3009TokenRequesterInterface
	sut                         *x402_facilitator_services.PaymentVerifier
}

func WhenVerifyingAPaymentBeforeEach(t *testing.T, asset common.Address) *WhenVerifyingAPaymentTestingSuite {
	mockController := gomock.NewController(t)
	mockPayToEligibilityChecker := interfaces_mocks.NewMockPayToEligibilityCheckerInterface(mockController)
	mockEip3009TokenRequester := interfaces_mocks.NewMockEip3009TokenRequesterInterface(mockController)
	tokensSettings := &x402_facilitator_settings.TokensSettings{
		80002: {asset: {Eip712Name: "USDC", Eip712Version: "2"}},
	}

	return &WhenVerifyingAPaymentTestingSuite{
		mockPayToEligibilityChecker: mockPayToEligibilityChecker,
		mockEip3009TokenRequester:   mockEip3009TokenRequester,
		sut: x402_facilitator_services.NewPaymentVerifier(
			tokensSettings,
			mockPayToEligibilityChecker,
			mockEip3009TokenRequester,
		),
	}
}

func (suite *WhenVerifyingAPaymentTestingSuite) assertInvalid(t *testing.T, err error, reason string) {
	var invalidPayment *apperrors.InvalidPaymentError
	require.ErrorAs(t, err, &invalidPayment)
	assert.Equal(t, reason, invalidPayment.Reason)
}

func (suite *WhenVerifyingAPaymentTestingSuite) assertUnavailable(
	t *testing.T,
	err error,
	code string,
	sourceErr error,
) {
	var unavailableErr *apperrors.UnavailableError
	require.ErrorAs(t, err, &unavailableErr)
	assert.Equal(t, code, unavailableErr.Code)
	assert.ErrorIs(t, err, sourceErr)
}

func (suite *WhenVerifyingAPaymentTestingSuite) expectEligible(payment *models.Payment) {
	suite.mockPayToEligibilityChecker.EXPECT().Check(gomock.Any(), payment.PayTo).Return(nil)
}

func (suite *WhenVerifyingAPaymentTestingSuite) expectUnusedNonce(payment *models.Payment) {
	suite.mockEip3009TokenRequester.EXPECT().
		AuthorizationState(
			gomock.Any(),
			uint64(80002),
			payment.Asset,
			payment.Authorization.From,
			payment.Authorization.Nonce,
		).
		Return(false, nil)
}

func (suite *WhenVerifyingAPaymentTestingSuite) expectBalance(payment *models.Payment, balance *big.Int) {
	suite.mockEip3009TokenRequester.EXPECT().
		BalanceOf(gomock.Any(), uint64(80002), payment.Asset, payment.Authorization.From).
		Return(balance, nil)
}

func (suite *WhenVerifyingAPaymentTestingSuite) expectSimulation(payment *models.Payment, err error) {
	suite.mockEip3009TokenRequester.EXPECT().
		SimulateTransferWithAuthorization(gomock.Any(), uint64(80002), payment.Asset, payment.Authorization).
		Return(err)
}

func TestWhenVerifyingAPayment(t *testing.T) {
	t.Parallel()

	asset := common_tests.RandomEthAddress()
	builder := func() *model_builders.PaymentRequestBuilder {
		return model_builders.NewPaymentRequestBuilder().WithAsset(asset)
	}

	t.Run("Given the network has no configured token", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithChainId(137).BuildPayment()

		t.Run("Should refuse the network", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidNetwork.String())
		})
	})

	t.Run("Given the asset is not a configured token", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidPaymentRequirements.String())
		})
	})

	t.Run("Given the payer signed under another EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithEip712Domain("USD Coin", "2").BuildPayment()

		t.Run("Should refuse the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidSignature.String())
		})
	})

	t.Run("Given the payer signed with a v of 0 or 1", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithRawRecoveryId().BuildPayment()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.expectUnusedNonce(payment)
			suite.expectBalance(payment, payment.Amount)
			suite.expectSimulation(payment, nil)
		}

		t.Run("Should accept the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the authorization is not valid yet", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithValidAfter(time.Now().Unix() + 120).BuildPayment()

		t.Run("Should refuse validAfter", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidAuthorizationValidAfter.String())
		})
	})

	t.Run("Given the authorization expires before a settlement could land", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithValidBefore(time.Now().Unix() + 2).BuildPayment()

		t.Run("Should refuse validBefore", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidAuthorizationValidBefore.String())
		})
	})

	t.Run("Given payTo is not eligible", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()
		notEligibleErr := apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"payTo is not eligible",
		)

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.mockPayToEligibilityChecker.EXPECT().Check(gomock.Any(), payment.PayTo).Return(notEligibleErr)
		}

		t.Run("Should return the eligibility error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			assert.Same(t, notEligibleErr, err)
		})
	})

	t.Run("Given the authorization state cannot be read", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()
		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.mockEip3009TokenRequester.EXPECT().
				AuthorizationState(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, sourceErr)
		}

		t.Run("Should return chain_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertUnavailable(t, err, "chain_unavailable", sourceErr)
		})
	})

	t.Run("Given the authorization nonce is already used", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.mockEip3009TokenRequester.EXPECT().
				AuthorizationState(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, nil)
		}

		t.Run("Should refuse the transaction state", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidTransactionState.String())
		})
	})

	t.Run("Given the balance cannot be read", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()
		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.expectUnusedNonce(payment)
			suite.mockEip3009TokenRequester.EXPECT().
				BalanceOf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, sourceErr)
		}

		t.Run("Should return chain_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertUnavailable(t, err, "chain_unavailable", sourceErr)
		})
	})

	t.Run("Given the payer balance is below the amount", func(t *testing.T) {
		t.Parallel()

		payment := builder().WithAmount("1000").BuildPayment()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.expectUnusedNonce(payment)
			suite.expectBalance(payment, big.NewInt(999))
		}

		t.Run("Should refuse for insufficient funds", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInsufficientFunds.String())
		})
	})

	t.Run("Given the transfer simulation reverts", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.expectUnusedNonce(payment)
			suite.expectBalance(payment, payment.Amount)
			suite.expectSimulation(payment, fmt.Errorf("%w: execution reverted", apperrors.ErrTransferRejected))
		}

		t.Run("Should refuse the transaction state", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertInvalid(t, err, enums.PaymentReasonInvalidTransactionState.String())
		})
	})

	t.Run("Given the relayer wallet cannot be loaded", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()
		sourceErr := fmt.Errorf("%w: vault sealed", apperrors.ErrRelayerUnavailable)

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.expectUnusedNonce(payment)
			suite.expectBalance(payment, payment.Amount)
			suite.expectSimulation(payment, sourceErr)
		}

		t.Run("Should return relayer_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertUnavailable(t, err, "relayer_unavailable", sourceErr)
		})
	})

	t.Run("Given the transfer simulation cannot run", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()
		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.expectUnusedNonce(payment)
			suite.expectBalance(payment, payment.Amount)
			suite.expectSimulation(payment, sourceErr)
		}

		t.Run("Should return chain_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			suite.assertUnavailable(t, err, "chain_unavailable", sourceErr)
		})
	})

	t.Run("Given every check passes", func(t *testing.T) {
		t.Parallel()

		payment := builder().BuildPayment()

		initSuite := func(suite *WhenVerifyingAPaymentTestingSuite) {
			suite.expectEligible(payment)
			suite.expectUnusedNonce(payment)
			suite.expectBalance(payment, payment.Amount)
			suite.expectSimulation(payment, nil)
		}

		t.Run("Should accept the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAPaymentBeforeEach(t, asset)
			initSuite(suite)

			err := suite.sut.Verify(context.Background(), payment)

			assert.NoError(t, err)
		})
	})
}
