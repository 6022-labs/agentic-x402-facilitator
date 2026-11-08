package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReadingATransferStatusTestingSuite struct {
	chain *fixtures.IntegrationChain
	sut   *blockchain_services.Eip3009TransferSubmitter
}

func WhenReadingATransferStatusBeforeEach(t *testing.T) *WhenReadingATransferStatusTestingSuite {
	chain := fixtures.NewIntegrationChain(
		t,
		blockchain_settings.ChainConfig{Confirmations: 1, ReceiptTimeout: 30 * time.Second},
	)
	mockRelayerMetrics := blockchain_metrics_mocks.NewMockRelayerMetricsInterface(gomock.NewController(t))
	mockRelayerMetrics.EXPECT().
		RecordGasSpent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	return &WhenReadingATransferStatusTestingSuite{
		chain: chain,
		sut: blockchain_services.NewEip3009TransferSubmitter(
			chain.EthClientFactory,
			chain.TransactorFactory,
			chain.ChainsSettings,
			chain.RedisLockStorage,
			chain.RedisStorage,
			mockRelayerMetrics,
			zap.NewNop(),
		),
	}
}

func TestWhenReadingATransferStatus(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingATransferStatusBeforeEach(t)

			_, err := suite.sut.TransferStatus(context.Background(), 1, common.HexToHash("0xabc"))

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the chain does not know the transaction", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer pending", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingATransferStatusBeforeEach(t)

			transfer, err := suite.sut.TransferStatus(
				context.Background(),
				suite.chain.ChainId,
				common.HexToHash("0xabc"),
			)

			require.NoError(t, err)
			assert.Equal(t, models.TransferStatusPending, transfer.Status)
		})
	})

	t.Run("Given a mined transfer short of its confirmations", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer pending", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingATransferStatusBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()
			txHash, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization,
			)
			require.NoError(t, err)
			_, err = suite.sut.AwaitTransfer(context.Background(), suite.chain.ChainId, txHash)
			require.NoError(t, err)
			suite.chain.ChainsSettings.Get(suite.chain.ChainId).Confirmations = 50

			transfer, err := suite.sut.TransferStatus(context.Background(), suite.chain.ChainId, txHash)

			require.NoError(t, err)
			assert.Equal(t, models.TransferStatusPending, transfer.Status)
		})
	})

	t.Run("Given a mined transfer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer confirmed", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingATransferStatusBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()
			txHash, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization,
			)
			require.NoError(t, err)
			_, err = suite.sut.AwaitTransfer(context.Background(), suite.chain.ChainId, txHash)
			require.NoError(t, err)

			transfer, err := suite.sut.TransferStatus(context.Background(), suite.chain.ChainId, txHash)

			require.NoError(t, err)
			assert.Equal(t, models.TransferStatusConfirmed, transfer.Status)
		})
	})
}
