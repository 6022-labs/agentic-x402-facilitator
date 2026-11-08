package configurations

import (
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/metrics/interfaces"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services/interfaces"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"go.uber.org/dig"
)

func AddX402FacilitatorBlockchainConfiguration(container *dig.Container) {
	if err := container.Provide(blockchain_settings.NewChainsSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(blockchain_settings.NewAgentCollectionsManagersSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewBlockchainRpcMetrics,
		dig.As(new(metrics_interfaces.BlockchainRpcMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewRelayerMetrics,
		dig.As(new(metrics_interfaces.RelayerMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(factory.NewEthClientFactory); err != nil {
		panic(err)
	}

	if err := container.Provide(factory.NewTransactorFactory); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewRelayerBalanceReader,
		dig.As(new(blockchain_interfaces.RelayerBalanceReaderInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewRelayerAccount,
		dig.As(new(x402_facilitator_interfaces.RelayerAccountInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewEip3009TokenRequester,
		dig.As(new(x402_facilitator_interfaces.Eip3009TokenRequesterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewEip3009TransferSubmitter,
		dig.As(new(x402_facilitator_interfaces.Eip3009TransferSubmitterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewAgentRegistryRequester,
		dig.As(new(x402_facilitator_interfaces.AgentRegistryRequesterInterface)),
	); err != nil {
		panic(err)
	}
}
