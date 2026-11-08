package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReadingTheRelayerBalancesTestingSuite struct {
	chain *fixtures.IntegrationChain
	sut   *blockchain_services.RelayerBalanceReader
}

// WhenReadingTheRelayerBalancesBeforeEach uses the chain's relayer unless relayerKeyStorage is given.
func WhenReadingTheRelayerBalancesBeforeEach(
	t *testing.T,
	relayerKeyStorage x402_facilitator_interfaces.RelayerKeyStorageInterface,
) *WhenReadingTheRelayerBalancesTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	transactorFactory := chain.TransactorFactory
	if relayerKeyStorage != nil {
		transactorFactory = factory.NewTransactorFactory(relayerKeyStorage, zap.NewNop())
	}
	(*chain.ChainsSettings)[unreachableChainId] = &blockchain_settings.ChainConfig{
		RpcHttpUrl:     "http://127.0.0.1:1",
		RequestTimeout: 2 * time.Second,
	}

	return &WhenReadingTheRelayerBalancesTestingSuite{
		chain: chain,
		sut: blockchain_services.NewRelayerBalanceReader(
			chain.EthClientFactory,
			transactorFactory,
			chain.ChainsSettings,
			zap.NewNop(),
		),
	}
}

func TestWhenReadingTheRelayerBalances(t *testing.T) {
	t.Parallel()

	t.Run("Given the relayer key cannot be read", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return no balance", func(t *testing.T) {
			t.Parallel()

			mockRelayerKeyStorage := interfaces_mocks.NewMockRelayerKeyStorageInterface(gomock.NewController(t))
			mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("", errors.New("vault sealed"))
			suite := WhenReadingTheRelayerBalancesBeforeEach(t, mockRelayerKeyStorage)

			assert.Empty(t, suite.sut.Balances(context.Background()))
		})
	})

	t.Run("Given one chain is reachable and another is not", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return only the reachable chain's balance", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerBalancesBeforeEach(t, nil)
			expected, err := suite.chain.Client.BalanceAt(context.Background(), suite.chain.Relayer.Address, nil)
			require.NoError(t, err)

			balances := suite.sut.Balances(context.Background())

			assert.Equal(t, expected, balances[suite.chain.ChainId])
			assert.NotContains(t, balances, unreachableChainId)
		})
	})
}
