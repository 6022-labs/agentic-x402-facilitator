package configurations

import (
	"github.com/6022-labs/agentic-x402-facilitator/src/common/settings"
	x402_facilitator_configurations "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/configurations"
	blockchain_configurations "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/configurations"
	mvc_configurations "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_mvc/configurations"
	redis_configurations "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/configurations"
	vault_configurations "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/configurations"
	"github.com/knadh/koanf/v2"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const (
	hostSettingsKey = "host"
	defaultApiPort  = 5000
)

func ConfigureDI(config *koanf.Koanf) *dig.Container {
	container := dig.New()

	if err := container.Provide(func(logger *zap.Logger, k *koanf.Koanf) *settings.HostSettings {
		return settings.NewHostSettings(logger, k, hostSettingsKey, defaultApiPort)
	}); err != nil {
		panic(err)
	}
	if err := container.Provide(func() *koanf.Koanf {
		return config
	}); err != nil {
		panic(err)
	}

	vault_configurations.AddX402FacilitatorVaultConfiguration(container)
	redis_configurations.AddX402FacilitatorRedisConfiguration(container)
	blockchain_configurations.AddX402FacilitatorBlockchainConfiguration(container)
	x402_facilitator_configurations.AddX402FacilitatorConfiguration(container)
	mvc_configurations.AddX402FacilitatorMvcConfiguration(container)

	return container
}
