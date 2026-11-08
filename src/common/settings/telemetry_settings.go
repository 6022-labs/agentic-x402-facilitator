package settings

import (
	"encoding/base64"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const TelemetrySettingsKey = "telemetry"

const (
	TelemetryAuthSchemeNone   = ""
	TelemetryAuthSchemeBearer = "bearer"
	TelemetryAuthSchemeBasic  = "basic"
)

// TelemetrySettings targets an OTLP collector; AuthScheme is derived from which credential is set.
type TelemetrySettings struct {
	Enabled      bool   `koanf:"enabled"`
	Endpoint     string `koanf:"endpoint"`
	Insecure     bool   `koanf:"insecure"`
	InstanceId   string `koanf:"instance_id"`
	Environment  string `koanf:"environment"`
	AuthToken    string `koanf:"auth_token"`
	AuthUsername string `koanf:"auth_username"`
	AuthPassword string `koanf:"auth_password"`
	AuthScheme   string `koanf:"-"`
}

func NewTelemetrySettings(
	logger *zap.Logger,
	k *koanf.Koanf,
) *TelemetrySettings {
	logger.Debug("loading settings", zap.String("key", TelemetrySettingsKey))

	var settings TelemetrySettings
	if err := k.UnmarshalWithConf(TelemetrySettingsKey, &settings, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal telemetry settings", zap.Error(err))
	}

	settings.Endpoint = strings.TrimSpace(settings.Endpoint)
	settings.InstanceId = strings.TrimSpace(settings.InstanceId)
	settings.Environment = strings.TrimSpace(settings.Environment)
	settings.AuthToken = strings.TrimSpace(settings.AuthToken)
	settings.AuthUsername = strings.TrimSpace(settings.AuthUsername)
	settings.AuthPassword = strings.TrimSpace(settings.AuthPassword)

	if settings.Enabled {
		if settings.Endpoint == "" {
			logger.Fatal("telemetry.endpoint is required when telemetry.enabled is true")
		}
		if settings.InstanceId == "" {
			logger.Fatal("telemetry.instance_id is required when telemetry.enabled is true")
		}
		if settings.Environment == "" {
			logger.Fatal("telemetry.environment is required when telemetry.enabled is true")
		}

		settings.resolveAuth(logger)
	}

	logger.Debug("loaded settings", zap.String("key", TelemetrySettingsKey))
	return &settings
}

// AuthorizationHeader returns the value the OTLP exporters must send, false when there is none.
func (s *TelemetrySettings) AuthorizationHeader() (string, bool) {
	switch s.AuthScheme {
	case TelemetryAuthSchemeBearer:
		return "Bearer " + s.AuthToken, true
	case TelemetryAuthSchemeBasic:
		credentials := s.AuthUsername + ":" + s.AuthPassword
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials)), true
	default:
		return "", false
	}
}

func (s *TelemetrySettings) resolveAuth(logger *zap.Logger) {
	hasBasic := s.AuthUsername != "" || s.AuthPassword != ""

	switch {
	case s.AuthToken != "" && hasBasic:
		logger.Fatal(
			"telemetry.auth_token and telemetry.auth_username/auth_password are mutually exclusive",
		)
	case s.AuthToken != "":
		s.AuthScheme = TelemetryAuthSchemeBearer
	case hasBasic:
		if s.AuthUsername == "" || s.AuthPassword == "" {
			logger.Fatal(
				"telemetry.auth_username and telemetry.auth_password must both be set " +
					"(TELEMETRY__AUTH_USERNAME, TELEMETRY__AUTH_PASSWORD)",
			)
		}
		s.AuthScheme = TelemetryAuthSchemeBasic
	}
}
