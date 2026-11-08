package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenCheckingPayToEligibilityTestingSuite struct {
	mockAgentRegistryRequester *interfaces_mocks.MockAgentRegistryRequesterInterface
	mockCacheStorage           *interfaces_mocks.MockCacheStorageInterface
	sut                        *x402_facilitator_services.PayToEligibilityChecker
}

func WhenCheckingPayToEligibilityBeforeEach(t *testing.T, agentsOnly bool) *WhenCheckingPayToEligibilityTestingSuite {
	mockController := gomock.NewController(t)
	mockAgentRegistryRequester := interfaces_mocks.NewMockAgentRegistryRequesterInterface(mockController)
	mockCacheStorage := interfaces_mocks.NewMockCacheStorageInterface(mockController)
	mockAgentRegistryRequester.EXPECT().RegistryChainIds().Return([]uint64{137, 80002}).AnyTimes()
	accessSettings := &x402_facilitator_settings.AccessSettings{
		AgentsOnly: agentsOnly,
		CacheTtl:   5 * time.Minute,
	}

	return &WhenCheckingPayToEligibilityTestingSuite{
		mockAgentRegistryRequester: mockAgentRegistryRequester,
		mockCacheStorage:           mockCacheStorage,
		sut: x402_facilitator_services.NewPayToEligibilityChecker(
			accessSettings,
			mockAgentRegistryRequester,
			mockCacheStorage,
			zap.NewNop(),
		),
	}
}

func (suite *WhenCheckingPayToEligibilityTestingSuite) cacheKey(payTo common.Address) string {
	return "x402:eligibility:137,80002:" + strings.ToLower(payTo.Hex())
}

func (suite *WhenCheckingPayToEligibilityTestingSuite) expectCached(payTo common.Address, value *string) {
	suite.mockCacheStorage.EXPECT().ReadString(gomock.Any(), suite.cacheKey(payTo)).Return(value, nil)
}

func (suite *WhenCheckingPayToEligibilityTestingSuite) assertNotEligible(t *testing.T, err error) {
	var invalidPayment *apperrors.InvalidPaymentError
	require.ErrorAs(t, err, &invalidPayment)
	assert.Equal(t, enums.PaymentReasonInvalidPaymentRequirements.String(), invalidPayment.Reason)
	assert.Equal(t, "payTo is not a registered 6022 agent address", invalidPayment.Message)
}

func TestWhenCheckingPayToEligibility(t *testing.T) {
	t.Parallel()

	eligible := "1"
	notEligible := "0"

	t.Run("Given access is open to every payee", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept payTo without any lookup", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, false)

			err := suite.sut.Check(context.Background(), common_tests.RandomEthAddress())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the store cannot be read", func(t *testing.T) {
		t.Parallel()

		payTo := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenCheckingPayToEligibilityTestingSuite) {
			suite.mockCacheStorage.EXPECT().ReadString(gomock.Any(), suite.cacheKey(payTo)).
				Return(nil, errors.New("redis down"))
			suite.mockCacheStorage.EXPECT().Store(gomock.Any(), suite.cacheKey(payTo), "1", gomock.Any()).Return(nil)
			suite.mockAgentRegistryRequester.EXPECT().IsAgentAddress(gomock.Any(), payTo).Return(true, nil)
		}

		t.Run("Should look payTo up in the registry", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payTo)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the store knows payTo is an agent address", func(t *testing.T) {
		t.Parallel()

		payTo := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenCheckingPayToEligibilityTestingSuite) {
			suite.expectCached(payTo, &eligible)
		}

		t.Run("Should accept it without reading the registry", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payTo)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the store knows payTo is not an agent address", func(t *testing.T) {
		t.Parallel()

		payTo := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenCheckingPayToEligibilityTestingSuite) {
			suite.expectCached(payTo, &notEligible)
		}

		t.Run("Should refuse the requirements without reading the registry", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payTo)

			suite.assertNotEligible(t, err)
		})
	})

	t.Run("Given the registry cannot be read", func(t *testing.T) {
		t.Parallel()

		payTo := common_tests.RandomEthAddress()
		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenCheckingPayToEligibilityTestingSuite) {
			suite.expectCached(payTo, nil)
			suite.mockAgentRegistryRequester.EXPECT().IsAgentAddress(gomock.Any(), payTo).Return(false, sourceErr)
		}

		t.Run("Should return registry_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payTo)

			var unavailableErr *apperrors.UnavailableError
			require.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "registry_unavailable", unavailableErr.Code)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the store cannot save the lookup", func(t *testing.T) {
		t.Parallel()

		payTo := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenCheckingPayToEligibilityTestingSuite) {
			suite.expectCached(payTo, nil)
			suite.mockAgentRegistryRequester.EXPECT().IsAgentAddress(gomock.Any(), payTo).Return(true, nil)
			suite.mockCacheStorage.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("redis down"))
		}

		t.Run("Should still answer from the registry", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)

			err := suite.sut.Check(context.Background(), payTo)

			assert.NoError(t, err)
		})
	})

	t.Run("Given payTo is not a registered agent address", func(t *testing.T) {
		t.Parallel()

		payTo := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenCheckingPayToEligibilityTestingSuite) {
			suite.expectCached(payTo, nil)
			suite.mockAgentRegistryRequester.EXPECT().IsAgentAddress(gomock.Any(), payTo).Return(false, nil)
		}

		t.Run("Should refuse the requirements", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)
			suite.mockCacheStorage.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := suite.sut.Check(context.Background(), payTo)

			suite.assertNotEligible(t, err)
		})

		t.Run("Should cache the refusal for the access ttl", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)
			suite.mockCacheStorage.EXPECT().Store(gomock.Any(), suite.cacheKey(payTo), "0", 5*time.Minute).Return(nil)

			_ = suite.sut.Check(context.Background(), payTo)
		})
	})

	t.Run("Given payTo is a registered agent address", func(t *testing.T) {
		t.Parallel()

		payTo := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenCheckingPayToEligibilityTestingSuite) {
			suite.expectCached(payTo, nil)
			suite.mockAgentRegistryRequester.EXPECT().IsAgentAddress(gomock.Any(), payTo).Return(true, nil)
		}

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)
			suite.mockCacheStorage.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := suite.sut.Check(context.Background(), payTo)

			assert.NoError(t, err)
		})

		t.Run("Should cache it for the access ttl", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingPayToEligibilityBeforeEach(t, true)
			initSuite(suite)
			suite.mockCacheStorage.EXPECT().Store(gomock.Any(), suite.cacheKey(payTo), "1", 5*time.Minute).Return(nil)

			_ = suite.sut.Check(context.Background(), payTo)
		})
	})
}
