package settings

import (
	"net"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const ProxySettingsKey = "proxy"

// ProxySettings resolves the client IP from Header, only on requests sent by TrustedProxies (IPs or CIDRs).
type ProxySettings struct {
	Header         string   `koanf:"header"`
	TrustedProxies []string `koanf:"trusted_proxies"`
}

func NewProxySettings(logger *zap.Logger, k *koanf.Koanf) *ProxySettings {
	// Not seeded: mapstructure merges into an existing slice element-wise.
	var settings ProxySettings
	if err := k.Unmarshal(ProxySettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal proxy settings", zap.Error(err))
	}

	settings.Header = strings.TrimSpace(settings.Header)
	settings.normalizeTrustedProxies(logger)
	if settings.Header != "" && len(settings.TrustedProxies) == 0 {
		logger.Warn("proxy.header ignored: proxy.trusted_proxies is empty")
		settings.Header = ""
	}

	return &settings
}

// Env hands the whole list through as one comma-separated string.
func (s *ProxySettings) normalizeTrustedProxies(logger *zap.Logger) {
	normalized := make([]string, 0, len(s.TrustedProxies))
	for _, entry := range s.TrustedProxies {
		for _, part := range strings.Split(entry, ",") {
			proxy := strings.TrimSpace(part)
			if proxy == "" {
				continue
			}
			if !s.isValidProxy(proxy) {
				logger.Fatal("proxy.trusted_proxies entries must be IPs or CIDR ranges", zap.String("proxy", proxy))
			}
			normalized = append(normalized, proxy)
		}
	}

	s.TrustedProxies = normalized
}

func (s *ProxySettings) isValidProxy(proxy string) bool {
	if net.ParseIP(proxy) != nil {
		return true
	}
	_, _, err := net.ParseCIDR(proxy)
	return err == nil
}
