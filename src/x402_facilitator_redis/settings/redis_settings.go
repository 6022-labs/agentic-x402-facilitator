package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RedisSettingsKey = "redis"

type RedisSettings struct {
	DB       int64  `koanf:"db"`
	Host     string `koanf:"host"`
	Password string `koanf:"password"`
}

func NewRedisSettings(logger *zap.Logger, k *koanf.Koanf) *RedisSettings {
	settings := RedisSettings{}
	if err := k.Unmarshal(RedisSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal redis settings", zap.Error(err))
	}

	settings.Host = strings.TrimSpace(settings.Host)
	settings.Password = strings.TrimSpace(settings.Password)

	if len(strings.TrimSpace(settings.Host)) == 0 {
		logger.Fatal("please set your redis.host value in your configuration")
	}

	if len(strings.TrimSpace(settings.Password)) == 0 {
		logger.Warn("redis password is not set")
	}

	if settings.DB < 0 {
		logger.Warn("invalid redis.db value, defaulting to 0")
		settings.DB = 0
	}

	return &settings
}
