package services_test

import (
	"context"
	"math/big"
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

type WhenAwaitingATransferTestingSuite struct {
	chain              *fixtures.IntegrationChain
	mockRelayerMetrics *blockchain_metrics_mocks.MockRelayerMetricsInterface
	sut                *blockchain_services.TransferSubmitter
}

func WhenAwaitingATransferBeforeEach(
	t *testing.T,
	chainConfig blockchain_settings.ChainConfig,
) *WhenAwaitingATransferTestingSuite {
	chain := fixtures.NewIntegrationChain(t, chainConfig)
	mockRelayerMetrics := blockchain_metrics_mocks.NewMockRelayerMetricsInterface(gomock.NewController(t))

	return &WhenAwaitingATransferTestingSuite{
		chain:              chain,
		mockRelayerMetrics: mockRelayerMetrics,
		sut: blockchain_services.NewTransferSubmitter(
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

func (suite *WhenAwaitingATransferTestingSuite) broadcast(t *testing.T, builder interface {
	BuildPayment() *models.Payment
}) common.Hash {
	payment := builder.BuildPayment()
	txHash, err := suite.sut.Broadcast(
		context.Background(),
		suite.chain.ChainId, suite.chain.TransferCall(t, payment))
	require.NoError(t, err)
	return txHash
}

func TestWhenAwaitingATransfer(t *testing.T) {
	t.Parallel()

	confirmedChain := blockchain_settings.ChainConfig{Confirmations: 1, ReceiptTimeout: 30 * time.Second}

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAwaitingATransferBeforeEach(t, confirmedChain)

			_, err := suite.sut.AwaitTransfer(context.Background(), 1, common.HexToHash("0xabc"))

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the chain does not reach the confirmations in time", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer pending", func(t *testing.T) {
			t.Parallel()

			suite := WhenAwaitingATransferBeforeEach(
				t,
				blockchain_settings.ChainConfig{Confirmations: 10_000, ReceiptTimeout: 3 * time.Second},
			)
			txHash := suite.broadcast(t, suite.chain.FundedPayment(t, 1000))

			transfer, err := suite.sut.AwaitTransfer(context.Background(), suite.chain.ChainId, txHash)

			require.NoError(t, err)
			assert.Equal(t, models.TransferStatusPending, transfer.Status)
		})
	})

	t.Run("Given a valid funded authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer confirmed", func(t *testing.T) {
			t.Parallel()

			suite := WhenAwaitingATransferBeforeEach(t, confirmedChain)
			suite.mockRelayerMetrics.EXPECT().
				RecordGasSpent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			txHash := suite.broadcast(t, suite.chain.FundedPayment(t, 1000))

			transfer, err := suite.sut.AwaitTransfer(context.Background(), suite.chain.ChainId, txHash)

			require.NoError(t, err)
			assert.Equal(t, models.TransferStatusConfirmed, transfer.Status)
		})

		t.Run("Should move the funds to payTo", func(t *testing.T) {
			t.Parallel()

			suite := WhenAwaitingATransferBeforeEach(t, confirmedChain)
			suite.mockRelayerMetrics.EXPECT().
				RecordGasSpent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			builder := suite.chain.FundedPayment(t, 1500)
			payment := builder.BuildPayment()
			txHash := suite.broadcast(t, builder)

			_, err := suite.sut.AwaitTransfer(context.Background(), suite.chain.ChainId, txHash)

			require.NoError(t, err)
			payerBalance, err := suite.chain.Token.BalanceOf(builder.Payer())
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(500), payerBalance)
			payeeBalance, err := suite.chain.Token.BalanceOf(payment.PayTo)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(1000), payeeBalance)
		})

		t.Run("Should record the gas the relayer paid", func(t *testing.T) {
			t.Parallel()

			suite := WhenAwaitingATransferBeforeEach(t, confirmedChain)
			suite.mockRelayerMetrics.EXPECT().
				RecordGasSpent(
					gomock.Any(),
					suite.chain.ChainId,
					gomock.Cond(func(gasUsed uint64) bool { return gasUsed > 21000 }),
					gomock.Cond(func(cost *big.Int) bool { return cost.Sign() > 0 }),
					"confirmed",
				)
			txHash := suite.broadcast(t, suite.chain.FundedPayment(t, 1000))

			_, err := suite.sut.AwaitTransfer(context.Background(), suite.chain.ChainId, txHash)

			require.NoError(t, err)
		})
	})
}
