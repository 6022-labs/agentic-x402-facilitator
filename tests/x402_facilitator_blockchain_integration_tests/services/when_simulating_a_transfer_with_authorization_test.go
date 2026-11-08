package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenSimulatingATransferWithAuthorizationTestingSuite struct {
	chain *fixtures.IntegrationChain
	sut   *blockchain_services.Eip3009TokenRequester
}

func WhenSimulatingATransferWithAuthorizationBeforeEach(
	t *testing.T,
) *WhenSimulatingATransferWithAuthorizationTestingSuite {
	chain := fixtures.NewIntegrationChain(t, blockchain_settings.ChainConfig{})
	return &WhenSimulatingATransferWithAuthorizationTestingSuite{
		chain: chain,
		sut: blockchain_services.NewEip3009TokenRequester(
			zap.NewNop(),
			chain.EthClientFactory,
			chain.TransactorFactory,
		),
	}
}

func TestWhenSimulatingATransferWithAuthorization(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferWithAuthorizationBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			err := suite.sut.SimulateTransferWithAuthorization(
				context.Background(),
				1,
				suite.chain.Token.Address,
				payment.Authorization,
			)

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the relayer key cannot be read", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a relayer unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferWithAuthorizationBeforeEach(t)
			mockRelayerKeyStorage := interfaces_mocks.NewMockRelayerKeyStorageInterface(gomock.NewController(t))
			mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("", errors.New("vault sealed"))
			sut := blockchain_services.NewEip3009TokenRequester(
				zap.NewNop(),
				suite.chain.EthClientFactory,
				factory.NewTransactorFactory(mockRelayerKeyStorage, zap.NewNop()),
			)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			err := sut.SimulateTransferWithAuthorization(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization,
			)

			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
		})
	})

	t.Run("Given the payer balance is below the amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer rejected", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferWithAuthorizationBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 10).BuildPayment()

			err := suite.sut.SimulateTransferWithAuthorization(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization,
			)

			assert.ErrorIs(t, err, apperrors.ErrTransferRejected)
		})
	})

	t.Run("Given the payer signed under another EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer rejected", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferWithAuthorizationBeforeEach(t)
			payment := suite.chain.FundedPayment(t, 1000).WithEip712Domain("USDC", "1").BuildPayment()

			err := suite.sut.SimulateTransferWithAuthorization(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization,
			)

			assert.ErrorIs(t, err, apperrors.ErrTransferRejected)
		})
	})

	t.Run("Given a valid funded authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should succeed without moving funds", func(t *testing.T) {
			t.Parallel()

			suite := WhenSimulatingATransferWithAuthorizationBeforeEach(t)
			builder := suite.chain.FundedPayment(t, 1000)
			payment := builder.BuildPayment()

			err := suite.sut.SimulateTransferWithAuthorization(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.Token.Address,
				payment.Authorization,
			)

			assert.NoError(t, err)
			balance, balanceErr := suite.chain.Token.BalanceOf(builder.Payer())
			assert.NoError(t, balanceErr)
			assert.Equal(t, int64(1000), balance.Int64())
		})
	})
}
