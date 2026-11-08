package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RelayerSettingsKey = "relayer"

const (
	defaultRelayerVaultKey = "relayer"
	RelayerPrivateKeyField = "private_key"
)

// RelayerSettings locates the relayer key in Vault; the key itself never sits in config.
type RelayerSettings struct {
	VaultKey string `koanf:"vault_key"`
}

func NewRelayerSettings(logger *zap.Logger, k *koanf.Koanf) *RelayerSettings {
	settings := RelayerSettings{VaultKey: defaultRelayerVaultKey}
	if err := k.Unmarshal(RelayerSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal relayer settings", zap.Error(err))
	}

	settings.VaultKey = strings.TrimSpace(settings.VaultKey)
	if settings.VaultKey == "" {
		settings.VaultKey = defaultRelayerVaultKey
	}

	return &settings
}
