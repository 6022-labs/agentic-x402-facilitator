package transports_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/transports"
	metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWhenReadingTheRpcOrigin(t *testing.T) {
	t.Parallel()

	t.Run("Given an RPC URL with credentials, an API key path and a query", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep only the scheme and host", func(t *testing.T) {
			t.Parallel()

			sut, err := transports.NewRpcRoundTripper(
				nil,
				"https://user:pass@rpc.example:8545/v2/secret?key=other",
				metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t)),
				"80002",
			)
			require.NoError(t, err)

			assert.Equal(t, "https://rpc.example:8545", sut.Origin())
		})
	})
}
