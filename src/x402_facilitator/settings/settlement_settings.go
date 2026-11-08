package settings

import (
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const SettlementSettingsKey = "settlement"

const defaultSettlementLockTtlSeconds = 600

// SettlementSettings bounds the in-flight lock, freed on its own if its replica dies mid-settlement.
type SettlementSettings struct {
	LockTtl time.Duration
}

type settlementSettingsRaw struct {
	LockTtlSeconds int `koanf:"lock_ttl_seconds"`
}

func NewSettlementSettings(logger *zap.Logger, k *koanf.Koanf) *SettlementSettings {
	raw := settlementSettingsRaw{LockTtlSeconds: defaultSettlementLockTtlSeconds}
	if err := k.Unmarshal(SettlementSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal settlement settings", zap.Error(err))
	}

	if raw.LockTtlSeconds <= 0 {
		logger.Warn(
			"invalid settlement.lock_ttl_seconds, using default",
			zap.Int("default", defaultSettlementLockTtlSeconds),
		)
		raw.LockTtlSeconds = defaultSettlementLockTtlSeconds
	}

	return &SettlementSettings{LockTtl: time.Duration(raw.LockTtlSeconds) * time.Second}
}
