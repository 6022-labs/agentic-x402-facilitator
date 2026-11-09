package services_test

import (
	"context"
	"math/big"
	"testing"

	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenCheckingAPermit2NonceTestingSuite struct {
	chain     *fixtures.IntegrationChain
	submitter *blockchain_services.TransferSubmitter
	sut       *blockchain_services.Permit2Requester
}

func WhenCheckingAPermit2NonceBeforeEach(t *testing.T) *WhenCheckingAPermit2NonceTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	return &WhenCheckingAPermit2NonceTestingSuite{
		chain: chain,
		submitter: blockchain_services.NewTransferSubmitter(
			chain.EthClientFactory,
			chain.TransactorFactory,
			chain.ChainsSettings,
			chain.RedisLockStorage,
			chain.RedisStorage,
			blockchain_metrics_mocks.NewMockRelayerMetricsInterface(gomock.NewController(t)),
			zap.NewNop(),
		),
		sut: blockchain_services.NewPermit2Requester(chain.EthClientFactory),
	}
}

func TestWhenCheckingAPermit2Nonce(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAPermit2NonceBeforeEach(t)

			_, err := suite.sut.NonceUsed(context.Background(), 1, common_tests.RandomEthAddress(), big.NewInt(1))

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given a nonce the owner never used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it unused", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAPermit2NonceBeforeEach(t)
			payment := suite.chain.FundedPermit2Payment(t, 1000).BuildPayment()

			used, err := suite.sut.NonceUsed(
				context.Background(),
				suite.chain.ChainId,
				payment.Authorization.From,
				payment.Authorization.NonceInt(),
			)

			require.NoError(t, err)
			assert.False(t, used)
		})
	})

	t.Run("Given a nonce a settlement consumed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it used", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAPermit2NonceBeforeEach(t)
			payment := suite.chain.FundedPermit2Payment(t, 1000).BuildPayment()
			_, err := suite.submitter.Broadcast(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.TransferCall(t, payment),
			)
			require.NoError(t, err)

			used, err := suite.sut.NonceUsed(
				context.Background(),
				suite.chain.ChainId,
				payment.Authorization.From,
				payment.Authorization.NonceInt(),
			)

			require.NoError(t, err)
			assert.True(t, used)
		})
	})
}
