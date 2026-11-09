package services_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"

	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenReadingABalanceTestingSuite struct {
	chain *fixtures.IntegrationChain
	sut   *blockchain_services.TokenRequester
}

func WhenReadingABalanceBeforeEach(t *testing.T) *WhenReadingABalanceTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	return &WhenReadingABalanceTestingSuite{
		chain: chain,
		sut:   blockchain_services.NewTokenRequester(chain.EthClientFactory),
	}
}

func TestWhenReadingABalance(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingABalanceBeforeEach(t)

			_, err := suite.sut.BalanceOf(
				context.Background(),
				1,
				suite.chain.Token.Address,
				common_tests.RandomEthAddress(),
			)

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the owner holds tokens", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the balance", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingABalanceBeforeEach(t)
			owner := common_tests.RandomEthAddress()
			require.NoError(t, suite.chain.Token.Mint(owner, big.NewInt(4200)))

			balance, err := suite.sut.BalanceOf(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				owner,
			)

			require.NoError(t, err)
			assert.Equal(t, big.NewInt(4200), balance)
		})
	})
}
