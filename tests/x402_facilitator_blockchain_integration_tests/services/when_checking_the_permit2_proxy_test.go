package services_test

import (
	"context"
	"testing"

	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenCheckingThePermit2ProxyTestingSuite struct {
	chain *fixtures.IntegrationChain
	sut   *blockchain_services.Permit2Requester
}

func WhenCheckingThePermit2ProxyBeforeEach(t *testing.T) *WhenCheckingThePermit2ProxyTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	return &WhenCheckingThePermit2ProxyTestingSuite{
		chain: chain,
		sut:   blockchain_services.NewPermit2Requester(chain.EthClientFactory),
	}
}

func TestWhenCheckingThePermit2Proxy(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingThePermit2ProxyBeforeEach(t)

			_, err := suite.sut.ProxyDeployed(context.Background(), 1)

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the proxy is deployed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it deployed", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingThePermit2ProxyBeforeEach(t)
			require.NoError(t, fixtures.InstallPermit2(context.Background(), suite.chain.Client))

			deployed, err := suite.sut.ProxyDeployed(context.Background(), suite.chain.ChainId)

			require.NoError(t, err)
			assert.True(t, deployed)
		})
	})
}
