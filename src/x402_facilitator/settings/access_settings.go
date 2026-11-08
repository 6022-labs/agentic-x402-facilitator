package settings

import (
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const AccessSettingsKey = "access"

const defaultAccessCacheTtlSeconds = 300

// AccessSettings limits payTo to 6022 agent addresses on any configured chain when AgentsOnly is set.
type AccessSettings struct {
	AgentsOnly bool
	CacheTtl   time.Duration
}

type accessSettingsRaw struct {
	AgentsOnly      bool `koanf:"agents_only"`
	CacheTtlSeconds int  `koanf:"cache_ttl_seconds"`
}

func NewAccessSettings(logger *zap.Logger, k *koanf.Koanf) *AccessSettings {
	raw := accessSettingsRaw{
		AgentsOnly:      true,
		CacheTtlSeconds: defaultAccessCacheTtlSeconds,
	}
	if err := k.Unmarshal(AccessSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal access settings", zap.Error(err))
	}

	if raw.CacheTtlSeconds <= 0 {
		logger.Warn("invalid access.cache_ttl_seconds, using default", zap.Int("default", defaultAccessCacheTtlSeconds))
		raw.CacheTtlSeconds = defaultAccessCacheTtlSeconds
	}

	return &AccessSettings{
		AgentsOnly: raw.AgentsOnly,
		CacheTtl:   time.Duration(raw.CacheTtlSeconds) * time.Second,
	}
}
