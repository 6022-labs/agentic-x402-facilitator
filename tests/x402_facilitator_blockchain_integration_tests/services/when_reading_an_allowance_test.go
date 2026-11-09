package services_test

import (
	"context"
	"math/big"
	"testing"

	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenReadingAnAllowanceTestingSuite struct {
	chain *fixtures.IntegrationChain
	sut   *blockchain_services.TokenRequester
}

func WhenReadingAnAllowanceBeforeEach(t *testing.T) *WhenReadingAnAllowanceTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	return &WhenReadingAnAllowanceTestingSuite{
		chain: chain,
		sut:   blockchain_services.NewTokenRequester(chain.EthClientFactory),
	}
}

func TestWhenReadingAnAllowance(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnAllowanceBeforeEach(t)

			_, err := suite.sut.Allowance(
				context.Background(),
				1,
				suite.chain.Token.Address,
				common_tests.RandomEthAddress(),
				common_tests.RandomEthAddress(),
			)

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the owner approved the spender", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the allowance", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnAllowanceBeforeEach(t)
			owner := common_tests.RandomEthAddress()
			spender := common_tests.RandomEthAddress()
			require.NoError(t, suite.chain.Token.SetAllowance(owner, spender, big.NewInt(4200)))

			allowance, err := suite.sut.Allowance(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				owner,
				spender,
			)

			require.NoError(t, err)
			assert.Equal(t, big.NewInt(4200), allowance)
		})
	})
}
