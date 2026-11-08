package transports_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/transports"
	metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWhenBuildingAnRpcRoundTripper(t *testing.T) {
	t.Parallel()

	t.Run("Given an RPC URL that is not a URL", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error that does not quote it", func(t *testing.T) {
			t.Parallel()

			_, err := transports.NewRpcRoundTripper(
				nil,
				"not a url secret-api-key",
				metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t)),
				"80002",
			)

			assert.ErrorContains(t, err, "rpc url of chain 80002 is not a URL")
			assert.NotContains(t, err.Error(), "secret-api-key")
		})
	})
}
