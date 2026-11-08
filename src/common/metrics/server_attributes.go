package metrics

import (
	"net/url"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
)

var defaultPortsByScheme = map[string]int{"http": 80, "https": 443}

// ServerAttributes maps a url to semconv server.*; the port falls back to the scheme default, which semconv wants set.
func ServerAttributes(serverUrl string) []attribute.KeyValue {
	parsed, err := url.Parse(serverUrl)
	if err != nil || parsed.Hostname() == "" {
		return nil
	}

	attrs := []attribute.KeyValue{attribute.String("server.address", parsed.Hostname())}

	if port, err := strconv.Atoi(parsed.Port()); err == nil {
		return append(attrs, attribute.Int("server.port", port))
	}
	if port, found := defaultPortsByScheme[parsed.Scheme]; found {
		return append(attrs, attribute.Int("server.port", port))
	}
	return attrs
}
