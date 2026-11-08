package settings

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const defaultHostBodyLimitBytes = 4 * 1024 * 1024

type HostSettings struct {
	ApiPort        int    `koanf:"api_port"`
	BodyLimitBytes int    `koanf:"body_limit_bytes"`
	ListenAddress  string `koanf:"listen_address"` // empty = all interfaces
}

// NewHostSettings loads a host's server settings; key and defaultApiPort are passed by the host so this stays generic.
func NewHostSettings(logger *zap.Logger, k *koanf.Koanf, key string, defaultApiPort int) *HostSettings {
	settings := HostSettings{
		ApiPort:        defaultApiPort,
		BodyLimitBytes: defaultHostBodyLimitBytes,
	}
	if err := k.Unmarshal(key, &settings); err != nil {
		logger.Fatal("failed to unmarshal host settings", zap.String("key", key), zap.Error(err))
	}

	if settings.ApiPort <= 0 {
		logger.Warn(fmt.Sprintf("invalid %s.api_port, using default", key), zap.Int("default", defaultApiPort))
		settings.ApiPort = defaultApiPort
	}
	if settings.BodyLimitBytes <= 0 {
		logger.Warn(fmt.Sprintf("invalid %s.body_limit_bytes, using 4MB", key))
		settings.BodyLimitBytes = defaultHostBodyLimitBytes
	}

	return &settings
}
