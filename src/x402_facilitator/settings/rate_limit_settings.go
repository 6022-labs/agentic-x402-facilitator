package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RateLimitSettingsKey = "rate_limit"

type RateLimitSettings struct {
	RequestsPerSecond int `koanf:"requests_per_second"`
}

func NewRateLimitSettings(
	logger *zap.Logger,
	k *koanf.Koanf,
) *RateLimitSettings {
	settings := RateLimitSettings{RequestsPerSecond: 50}
	if err := k.Unmarshal(RateLimitSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal rate_limit settings", zap.Error(err))
	}

	if settings.RequestsPerSecond <= 0 {
		logger.Warn("invalid rate_limit.requests_per_second value, defaulting to 50")
		settings.RequestsPerSecond = 50
	}

	return &settings
}
