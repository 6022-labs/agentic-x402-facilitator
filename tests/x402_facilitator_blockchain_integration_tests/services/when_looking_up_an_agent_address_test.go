package services_test

import (
	"context"
	"testing"
	"time"

	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	// otherChainId stands for a second chain: same anvil node, its own manager.
	otherChainId = uint64(31338)
	// unreachableChainId points at a closed port: a chain whose RPC is down.
	unreachableChainId = uint64(31339)
)

type WhenLookingUpAnAgentAddressTestingSuite struct {
	firstManager  *fixtures.AgentCollectionsManager
	secondManager *fixtures.AgentCollectionsManager
	sut           *blockchain_services.AgentRegistryRequester
}

func WhenLookingUpAnAgentAddressBeforeEach(
	t *testing.T,
	agentsOnly bool,
	withUnreachableChain bool,
) *WhenLookingUpAnAgentAddressTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	firstManager, err := fixtures.DeployAgentCollectionsManager(
		context.Background(),
		chain.Client,
		container_builders.GetAnvilContainerBuilder().Create(),
	)
	require.NoError(t, err)
	secondManager, err := fixtures.DeployAgentCollectionsManager(
		context.Background(),
		chain.Client,
		container_builders.GetAnvilContainerBuilder().Create(),
	)
	require.NoError(t, err)

	chainsSettings := blockchain_settings.ChainsSettings{
		chain.ChainId: chain.ChainsSettings.Get(chain.ChainId),
		otherChainId:  {RpcHttpUrl: chain.Relayer.HttpUrl, RequestTimeout: 10 * time.Second},
	}
	managersSettings := blockchain_settings.AgentCollectionsManagersSettings{
		chain.ChainId: firstManager.Address,
		otherChainId:  secondManager.Address,
	}
	if withUnreachableChain {
		chainsSettings[unreachableChainId] = &blockchain_settings.ChainConfig{
			RpcHttpUrl:     "http://127.0.0.1:1",
			RequestTimeout: 2 * time.Second,
		}
		managersSettings[unreachableChainId] = common_tests.RandomEthAddress()
	}

	mockBlockchainRpcMetrics := blockchain_metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t))
	mockBlockchainRpcMetrics.EXPECT().
		RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	return &WhenLookingUpAnAgentAddressTestingSuite{
		firstManager:  firstManager,
		secondManager: secondManager,
		sut: blockchain_services.NewAgentRegistryRequester(
			zap.NewNop(),
			&x402_facilitator_settings.AccessSettings{AgentsOnly: agentsOnly},
			&chainsSettings,
			&managersSettings,
			factory.NewEthClientFactory(&chainsSettings, mockBlockchainRpcMetrics),
		),
	}
}

func TestWhenLookingUpAnAgentAddress(t *testing.T) {
	t.Parallel()

	t.Run("Given access is open to every payee", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a registry not bound error", func(t *testing.T) {
			t.Parallel()

			suite := WhenLookingUpAnAgentAddressBeforeEach(t, false, false)

			_, err := suite.sut.IsAgentAddress(context.Background(), common_tests.RandomEthAddress())

			assert.ErrorContains(t, err, "agent registry is not bound")
		})
	})

	t.Run("Given a chain is down and no other chain knows the address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenLookingUpAnAgentAddressBeforeEach(t, true, true)

			_, err := suite.sut.IsAgentAddress(context.Background(), common_tests.RandomEthAddress())

			assert.ErrorContains(t, err, "chain 31339")
		})
	})

	t.Run("Given a chain is down but another chain knows the address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report an agent address", func(t *testing.T) {
			t.Parallel()

			suite := WhenLookingUpAnAgentAddressBeforeEach(t, true, true)
			address := common_tests.RandomEthAddress()
			require.NoError(t, suite.firstManager.Register(address, 7))

			isAgent, err := suite.sut.IsAgentAddress(context.Background(), address)

			require.NoError(t, err)
			assert.True(t, isAgent)
		})
	})

	t.Run("Given no chain knows the address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no agent address", func(t *testing.T) {
			t.Parallel()

			suite := WhenLookingUpAnAgentAddressBeforeEach(t, true, false)

			isAgent, err := suite.sut.IsAgentAddress(context.Background(), common_tests.RandomEthAddress())

			require.NoError(t, err)
			assert.False(t, isAgent)
		})
	})

	t.Run("Given only another chain knows the address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report an agent address", func(t *testing.T) {
			t.Parallel()

			suite := WhenLookingUpAnAgentAddressBeforeEach(t, true, false)
			address := common_tests.RandomEthAddress()
			require.NoError(t, suite.secondManager.Register(address, 7))

			isAgent, err := suite.sut.IsAgentAddress(context.Background(), address)

			require.NoError(t, err)
			assert.True(t, isAgent)
		})
	})

	t.Run("Given the first chain knows the address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report an agent address", func(t *testing.T) {
			t.Parallel()

			suite := WhenLookingUpAnAgentAddressBeforeEach(t, true, false)
			address := common_tests.RandomEthAddress()
			require.NoError(t, suite.firstManager.Register(address, 7))

			isAgent, err := suite.sut.IsAgentAddress(context.Background(), address)

			require.NoError(t, err)
			assert.True(t, isAgent)
		})
	})
}
