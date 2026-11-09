package services_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenSimulatingATransferTestingSuite struct {
	chain *fixtures.IntegrationChain
	sut   *blockchain_services.TransferSubmitter
}

// WhenSimulatingATransferBeforeEach uses the chain's relayer unless transactorFactory is given.
func WhenSimulatingATransferBeforeEach(
	t *testing.T,
	transactorFactory *factory.TransactorFactory,
) *WhenSimulatingATransferTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	if transactorFactory == nil {
		transactorFactory = chain.TransactorFactory
	}
	return &WhenSimulatingATransferTestingSuite{
		chain: chain,
		sut: blockchain_services.NewTransferSubmitter(
			chain.EthClientFactory,
			transactorFactory,
			chain.ChainsSettings,
			chain.RedisLockStorage,
			chain.RedisStorage,
			blockchain_metrics_mocks.NewMockRelayerMetricsInterface(gomock.NewController(t)),
			zap.NewNop(),
		),
	}
}

func TestWhenSimulatingATransfer(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferBeforeEach(t, nil)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			err := suite.sut.Simulate(context.Background(), 1, suite.chain.TransferCall(t, payment))

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the relayer key cannot be read", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a relayer unavailable error", func(t *testing.T) {
			t.Parallel()

			mockRelayerKeyStorage := interfaces_mocks.NewMockRelayerKeyStorageInterface(gomock.NewController(t))
			mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("", errors.New("vault sealed"))
			suite := WhenSimulatingATransferBeforeEach(
				t,
				factory.NewTransactorFactory(mockRelayerKeyStorage, zap.NewNop()),
			)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			err := suite.sut.Simulate(context.Background(), suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
		})
	})

	t.Run("Given the payer balance is below the amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer rejected", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferBeforeEach(t, nil)
			payment := suite.chain.FundedPayment(t, 10).BuildPayment()

			err := suite.sut.Simulate(context.Background(), suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			assert.ErrorIs(t, err, apperrors.ErrTransferRejected)
		})
	})

	t.Run("Given the payer signed under another EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer rejected", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferBeforeEach(t, nil)
			payment := suite.chain.FundedPayment(t, 1000).WithEip712Domain("USDC", "1").BuildPayment()

			err := suite.sut.Simulate(context.Background(), suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			assert.ErrorIs(t, err, apperrors.ErrTransferRejected)
		})
	})

	t.Run("Given a funded EIP-3009 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should succeed without moving funds", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferBeforeEach(t, nil)
			builder := suite.chain.FundedPayment(t, 1000)

			err := suite.sut.Simulate(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.TransferCall(t, builder.BuildPayment()),
			)

			require.NoError(t, err)
			balance, err := suite.chain.Token.BalanceOf(builder.Payer())
			require.NoError(t, err)
			assert.Equal(t, int64(1000), balance.Int64())
		})
	})

	t.Run("Given a funded Permit2 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should succeed without moving funds", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferBeforeEach(t, nil)
			builder := suite.chain.FundedPermit2Payment(t, 1000)

			err := suite.sut.Simulate(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.TransferCall(t, builder.BuildPayment()),
			)

			require.NoError(t, err)
			balance, err := suite.chain.Token.BalanceOf(builder.Payer())
			require.NoError(t, err)
			assert.Equal(t, int64(1000), balance.Int64())
		})
	})
}
