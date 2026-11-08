package host_configurations

import (
	"errors"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const defaultConfigFileName = "appsettings.json"

func LoadKoanfConfig() (*koanf.Koanf, error) {
	return LoadKoanfConfigFromFile(defaultConfigFileName)
}

func LoadKoanfConfigFromFile(path string) (*koanf.Koanf, error) {
	k := koanf.New(".")

	if len(strings.TrimSpace(path)) == 0 {
		path = defaultConfigFileName
	}

	if _, err := os.Stat(path); err == nil {
		if err := k.Load(file.Provider(path), json.Parser()); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err := loadEnvIntoKoanf(k); err != nil {
		return nil, err
	}

	return k, nil
}

func loadEnvIntoKoanf(k *koanf.Koanf) error {
	envMap := map[string]any{}
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]
		if !strings.Contains(key, "__") {
			continue
		}

		normalizedKey := strings.ToLower(strings.ReplaceAll(key, "__", "."))
		envMap[normalizedKey] = value
	}

	if len(envMap) == 0 {
		return nil
	}

	return k.Load(confmap.Provider(envMap, "."), nil)
}
