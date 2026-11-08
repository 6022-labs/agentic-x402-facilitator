package services_test

import (
	"testing"
	"time"

	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenListingRegistryChainIdsTestingSuite struct {
	sut *blockchain_services.AgentRegistryRequester
}

func WhenListingRegistryChainIdsBeforeEach(t *testing.T, agentsOnly bool) *WhenListingRegistryChainIdsTestingSuite {
	chainsSettings := &blockchain_settings.ChainsSettings{
		80002:    {RpcHttpUrl: "http://127.0.0.1:1", RequestTimeout: time.Second, Testnet: true},
		84532:    {RpcHttpUrl: "http://127.0.0.1:1", RequestTimeout: time.Second, Testnet: true},
		11155111: {RpcHttpUrl: "http://127.0.0.1:1", RequestTimeout: time.Second, Testnet: true},
	}
	managersSettings := &blockchain_settings.AgentCollectionsManagersSettings{
		11155111: common_tests.RandomEthAddress(),
		80002:    common_tests.RandomEthAddress(),
	}

	return &WhenListingRegistryChainIdsTestingSuite{
		sut: blockchain_services.NewAgentRegistryRequester(
			zap.NewNop(),
			&x402_facilitator_settings.AccessSettings{AgentsOnly: agentsOnly},
			chainsSettings,
			managersSettings,
			factory.NewEthClientFactory(
				chainsSettings,
				blockchain_metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t)),
			),
		),
	}
}

func TestWhenListingRegistryChainIds(t *testing.T) {
	t.Parallel()

	t.Run("Given access is open to every payee", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list no chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingRegistryChainIdsBeforeEach(t, false)

			assert.Empty(t, suite.sut.RegistryChainIds())
		})
	})

	t.Run("Given some configured chains have a registry", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list only those, sorted", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingRegistryChainIdsBeforeEach(t, true)

			assert.Equal(t, []uint64{80002, 11155111}, suite.sut.RegistryChainIds())
		})
	})
}
