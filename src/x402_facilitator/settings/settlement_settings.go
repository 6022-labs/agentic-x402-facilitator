package settings

import (
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const SettlementSettingsKey = "settlement"

const (
	defaultSettlementLockTtlSeconds = 600
	defaultMaxPerPayerPerHour       = 300
)

// SettlementSettings bounds the in-flight lock, freed on its own if its replica dies mid-settlement.
type SettlementSettings struct {
	LockTtl time.Duration
	// MaxPerPayerPerHour is per chain; 0 means no cap.
	MaxPerPayerPerHour int64
}

type settlementSettingsRaw struct {
	LockTtlSeconds     int   `koanf:"lock_ttl_seconds"`
	MaxPerPayerPerHour int64 `koanf:"max_per_payer_per_hour"`
}

func NewSettlementSettings(logger *zap.Logger, k *koanf.Koanf) *SettlementSettings {
	raw := settlementSettingsRaw{
		LockTtlSeconds:     defaultSettlementLockTtlSeconds,
		MaxPerPayerPerHour: defaultMaxPerPayerPerHour,
	}
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

	if raw.MaxPerPayerPerHour < 0 {
		logger.Warn(
			"invalid settlement.max_per_payer_per_hour, using default",
			zap.Int64("default", defaultMaxPerPayerPerHour),
		)
		raw.MaxPerPayerPerHour = defaultMaxPerPayerPerHour
	}

	return &SettlementSettings{
		LockTtl:            time.Duration(raw.LockTtlSeconds) * time.Second,
		MaxPerPayerPerHour: raw.MaxPerPayerPerHour,
	}
}
