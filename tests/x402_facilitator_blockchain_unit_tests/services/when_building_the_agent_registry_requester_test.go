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
	"go.uber.org/zap/zapcore"
)

type WhenBuildingTheAgentRegistryRequesterTestingSuite struct {
	chainsSettings   *blockchain_settings.ChainsSettings
	ethClientFactory *factory.EthClientFactory
}

func WhenBuildingTheAgentRegistryRequesterBeforeEach(
	t *testing.T,
	mainnetChainTestnet bool,
) *WhenBuildingTheAgentRegistryRequesterTestingSuite {
	chainsSettings := &blockchain_settings.ChainsSettings{
		137:   {RpcHttpUrl: "http://127.0.0.1:1", RequestTimeout: time.Second, Testnet: mainnetChainTestnet},
		80002: {RpcHttpUrl: "http://127.0.0.1:1", RequestTimeout: time.Second, Testnet: true},
	}
	return &WhenBuildingTheAgentRegistryRequesterTestingSuite{
		chainsSettings: chainsSettings,
		ethClientFactory: factory.NewEthClientFactory(
			chainsSettings,
			blockchain_metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t)),
		),
	}
}

func (suite *WhenBuildingTheAgentRegistryRequesterTestingSuite) panicLogger() *zap.Logger {
	return zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))
}

func TestWhenBuildingTheAgentRegistryRequester(t *testing.T) {
	t.Parallel()

	accessSettings := &x402_facilitator_settings.AccessSettings{AgentsOnly: true}

	t.Run("Given access is open to every payee", func(t *testing.T) {
		t.Parallel()

		managersSettings := &blockchain_settings.AgentCollectionsManagersSettings{
			137:   common_tests.RandomEthAddress(),
			80002: common_tests.RandomEthAddress(),
		}

		t.Run("Should build without binding a registry", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingTheAgentRegistryRequesterBeforeEach(t, false)

			assert.NotPanics(t, func() {
				blockchain_services.NewAgentRegistryRequester(
					suite.panicLogger(),
					&x402_facilitator_settings.AccessSettings{AgentsOnly: false},
					suite.chainsSettings,
					managersSettings,
					suite.ethClientFactory,
				)
			})
		})
	})

	t.Run("Given no configured chain has a registry", func(t *testing.T) {
		t.Parallel()

		managersSettings := &blockchain_settings.AgentCollectionsManagersSettings{}

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingTheAgentRegistryRequesterBeforeEach(t, true)

			assert.Panics(t, func() {
				blockchain_services.NewAgentRegistryRequester(
					suite.panicLogger(),
					accessSettings,
					suite.chainsSettings,
					managersSettings,
					suite.ethClientFactory,
				)
			})
		})
	})

	t.Run("Given a testnet chain and a mainnet chain", func(t *testing.T) {
		t.Parallel()

		managersSettings := &blockchain_settings.AgentCollectionsManagersSettings{
			80002: common_tests.RandomEthAddress(),
		}

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingTheAgentRegistryRequesterBeforeEach(t, false)

			assert.Panics(t, func() {
				blockchain_services.NewAgentRegistryRequester(
					suite.panicLogger(),
					accessSettings,
					suite.chainsSettings,
					managersSettings,
					suite.ethClientFactory,
				)
			})
		})
	})

	t.Run("Given chains of one kind of network", func(t *testing.T) {
		t.Parallel()

		managersSettings := &blockchain_settings.AgentCollectionsManagersSettings{
			80002: common_tests.RandomEthAddress(),
		}

		t.Run("Should build", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingTheAgentRegistryRequesterBeforeEach(t, true)

			assert.NotPanics(t, func() {
				blockchain_services.NewAgentRegistryRequester(
					suite.panicLogger(),
					accessSettings,
					suite.chainsSettings,
					managersSettings,
					suite.ethClientFactory,
				)
			})
		})
	})
}
