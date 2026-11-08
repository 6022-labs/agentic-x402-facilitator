package configurations

import (
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/metrics/interfaces"
	vault_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/services"
	vault_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/settings"
	vault "github.com/hashicorp/vault/api"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddX402FacilitatorVaultConfiguration(container *dig.Container) {
	if err := container.Provide(vault_settings.NewVaultSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(vault_settings.NewRelayerSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewVaultMetrics,
		dig.As(new(metrics_interfaces.VaultMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(provideVaultClient); err != nil {
		panic(err)
	}

	if err := container.Provide(
		vault_services.NewKvV1RelayerKeyStorage,
		dig.As(new(x402_facilitator_interfaces.RelayerKeyStorageInterface)),
	); err != nil {
		panic(err)
	}
}

func provideVaultClient(vaultSettings *vault_settings.VaultSettings, logger *zap.Logger) *vault.Client {
	cfg := vault.DefaultConfig()
	cfg.Address = vaultSettings.Address

	client, err := vault.NewClient(cfg)
	if err != nil {
		logger.Fatal("failed to create vault client", zap.Error(err))
	}
	client.SetToken(vaultSettings.Token)
	return client
}
