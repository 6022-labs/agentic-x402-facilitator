package host_configurations_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/common/host_configurations"
	"github.com/stretchr/testify/assert"
)

func TestWhenLoadingKoanfConfig(t *testing.T) {
	t.Run("Given a config file and a matching env override", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "appsettings.json")
		err := os.WriteFile(path, []byte(`{"host":{"api_port":5000},"redis":{"host":"from-file"}}`), 0o600)
		assert.NoError(t, err)

		t.Run("Should map double underscores to dots and let env win over the file", func(t *testing.T) {
			t.Setenv("VAULT__ADDRESS", "from-env")

			k, err := host_configurations.LoadKoanfConfigFromFile(path)

			assert.NoError(t, err)
			assert.Equal(t, "from-env", k.String("vault.address"))
			assert.Equal(t, 5000, k.Int("host.api_port"))
		})
	})

	t.Run("Given a nested env variable with no file", func(t *testing.T) {
		t.Run("Should lowercase the key and split it on the double underscore", func(t *testing.T) {
			t.Setenv("CHAINS__80002__RPC_HTTP_URL", "https://rpc.example")

			k, err := host_configurations.LoadKoanfConfigFromFile(filepath.Join(t.TempDir(), "missing.json"))

			assert.NoError(t, err)
			assert.Equal(t, "https://rpc.example", k.String("chains.80002.rpc_http_url"))
		})
	})

	t.Run("Given an env variable without a double underscore", func(t *testing.T) {
		t.Run("Should ignore it", func(t *testing.T) {
			t.Setenv("PATH_LIKE_VALUE", "ignored")

			k, err := host_configurations.LoadKoanfConfigFromFile(filepath.Join(t.TempDir(), "missing.json"))

			assert.NoError(t, err)
			assert.Equal(t, "", k.String("path_like_value"))
		})
	})
}
