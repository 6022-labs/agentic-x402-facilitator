package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const VaultSettingsKey = "vault"

type VaultSettings struct {
	Address       string `koanf:"address"`
	Token         string `koanf:"token"`
	KvV1MountPath string `koanf:"kv_v1_mount_path"`
}

func NewVaultSettings(logger *zap.Logger, k *koanf.Koanf) *VaultSettings {
	settings := VaultSettings{
		KvV1MountPath: "secret",
	}
	if err := k.Unmarshal(VaultSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal vault settings", zap.Error(err))
	}

	settings.Address = strings.TrimSpace(settings.Address)
	settings.Token = strings.TrimSpace(settings.Token)
	settings.KvV1MountPath = strings.TrimSpace(settings.KvV1MountPath)

	if len(settings.Address) == 0 {
		logger.Fatal("please set vault.address (or VAULT__ADDRESS env)")
	}
	if len(settings.Token) == 0 {
		logger.Fatal("please set vault.token (or VAULT__TOKEN env)")
	}
	if len(settings.KvV1MountPath) == 0 {
		settings.KvV1MountPath = "secret"
	}

	return &settings
}
