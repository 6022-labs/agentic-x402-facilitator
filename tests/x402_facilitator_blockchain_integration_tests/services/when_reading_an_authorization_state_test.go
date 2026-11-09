package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReadingAnAuthorizationStateTestingSuite struct {
	chain     *fixtures.IntegrationChain
	submitter *blockchain_services.TransferSubmitter
	sut       *blockchain_services.TokenRequester
}

func WhenReadingAnAuthorizationStateBeforeEach(t *testing.T) *WhenReadingAnAuthorizationStateTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{ReceiptTimeout: 30 * time.Second})
	mockRelayerMetrics := blockchain_metrics_mocks.NewMockRelayerMetricsInterface(gomock.NewController(t))
	mockRelayerMetrics.EXPECT().
		RecordGasSpent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	return &WhenReadingAnAuthorizationStateTestingSuite{
		chain: chain,
		submitter: blockchain_services.NewTransferSubmitter(
			chain.EthClientFactory,
			chain.TransactorFactory,
			chain.ChainsSettings,
			chain.RedisLockStorage,
			chain.RedisStorage,
			mockRelayerMetrics,
			zap.NewNop(),
		),
		sut: blockchain_services.NewTokenRequester(chain.EthClientFactory),
	}
}

func TestWhenReadingAnAuthorizationState(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnAuthorizationStateBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			_, err := suite.sut.AuthorizationState(
				context.Background(),
				1,
				suite.chain.Token.Address,
				payment.Authorization.From,
				payment.Authorization.Nonce,
			)

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the nonce was never used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it unused", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnAuthorizationStateBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			used, err := suite.sut.AuthorizationState(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization.From,
				payment.Authorization.Nonce,
			)

			require.NoError(t, err)
			assert.False(t, used)
		})
	})

	t.Run("Given the authorization was settled", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the nonce used", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnAuthorizationStateBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()
			txHash, err := suite.submitter.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))
			require.NoError(t, err)
			transfer, err := suite.submitter.AwaitTransfer(context.Background(), suite.chain.ChainId, txHash)
			require.NoError(t, err)
			require.Equal(t, models.TransferStatusConfirmed, transfer.Status)

			used, err := suite.sut.AuthorizationState(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization.From,
				payment.Authorization.Nonce,
			)

			require.NoError(t, err)
			assert.True(t, used)
		})
	})
}
