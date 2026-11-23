package services_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_integration_tests/fixtures"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenBroadcastingATransferTestingSuite struct {
	chain         *fixtures.IntegrationChain
	mockGasBudget *interfaces_mocks.MockGasBudgetInterface
	sut           *blockchain_services.TransferSubmitter
}

// WhenBroadcastingATransferBeforeEach uses the chain's relayer unless relayerKeyStorage is given.
func WhenBroadcastingATransferBeforeEach(
	t *testing.T,
	relayerKeyStorage x402_facilitator_interfaces.RelayerKeyStorageInterface,
) *WhenBroadcastingATransferTestingSuite {
	chain := fixtures.NewIntegrationChain(
		t,
		blockchain_settings.ChainConfig{Confirmations: 1, ReceiptTimeout: 30 * time.Second},
	)
	transactorFactory := chain.TransactorFactory
	if relayerKeyStorage != nil {
		transactorFactory = factory.NewTransactorFactory(relayerKeyStorage, zap.NewNop())
	}

	mockController := gomock.NewController(t)
	mockGasBudget := interfaces_mocks.NewMockGasBudgetInterface(mockController)

	return &WhenBroadcastingATransferTestingSuite{
		chain:         chain,
		mockGasBudget: mockGasBudget,
		sut: blockchain_services.NewTransferSubmitter(
			chain.EthClientFactory,
			transactorFactory,
			chain.ChainsSettings,
			chain.RedisLockStorage,
			chain.RedisStorage,
			mockGasBudget,
			blockchain_metrics_mocks.NewMockRelayerMetricsInterface(mockController),
			zap.NewNop(),
		),
	}
}

func (suite *WhenBroadcastingATransferTestingSuite) expectReserved() {
	suite.mockGasBudget.EXPECT().
		Reserve(
			gomock.Any(),
			suite.chain.ChainId,
			gomock.Any(),
			gomock.Cond(func(maxCost *big.Int) bool { return maxCost.Sign() > 0 }),
		).
		Return(nil)
}

func (suite *WhenBroadcastingATransferTestingSuite) relayerKey(suffix string) string {
	return fmt.Sprintf(
		"x402:relayer:%d:%s:%s",
		suite.chain.ChainId,
		strings.ToLower(suite.chain.Relayer.Address.Hex()),
		suffix,
	)
}

func TestWhenBroadcastingATransfer(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain has no configured RPC", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a no RPC configured error", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			_, err := suite.sut.Broadcast(context.Background(), 1, suite.chain.TransferCall(t, payment))

			assert.ErrorContains(t, err, "no RPC configured for chain 1")
		})
	})

	t.Run("Given the relayer key cannot be read", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a relayer unavailable error", func(t *testing.T) {
			t.Parallel()

			mockRelayerKeyStorage := interfaces_mocks.NewMockRelayerKeyStorageInterface(gomock.NewController(t))
			mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("", errors.New("vault sealed"))
			suite := WhenBroadcastingATransferBeforeEach(t, mockRelayerKeyStorage)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			_, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
		})
	})

	t.Run("Given another replica holds the send lock", func(t *testing.T) {
		t.Parallel()

		t.Run("Should give up after one lock ttl", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			chain := suite.chain.ChainsSettings.Get(suite.chain.ChainId)
			*chain = *blockchain_settings.NewChainConfig(
				chain.RpcHttpUrl,
				time.Second,
				chain.Confirmations,
				chain.ReceiptTimeout,
				chain.Testnet,
			)
			_, acquired, err := suite.chain.RedisLockStorage.TryLock(
				context.Background(),
				suite.relayerKey("lock"),
				time.Minute,
			)
			require.NoError(t, err)
			require.True(t, acquired)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			_, err = suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			assert.ErrorContains(t, err, "wait for relayer send lock")
		})
	})

	t.Run("Given a signed transaction was kept for the nonce gap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should resend it and take the next nonce", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			suite.expectReserved()
			pending, err := suite.chain.Client.PendingNonceAt(context.Background(), suite.chain.Relayer.Address)
			require.NoError(t, err)
			key, err := crypto.HexToECDSA(suite.chain.Relayer.PrivateKeyHex)
			require.NoError(t, err)
			lost, err := types.SignNewTx(
				key,
				types.LatestSignerForChainID(new(big.Int).SetUint64(suite.chain.ChainId)),
				&types.DynamicFeeTx{
					ChainID:   new(big.Int).SetUint64(suite.chain.ChainId),
					Nonce:     pending,
					GasTipCap: big.NewInt(1_000_000_000),
					GasFeeCap: big.NewInt(100_000_000_000),
					Gas:       21000,
					To:        &suite.chain.Relayer.Address,
				},
			)
			require.NoError(t, err)
			raw, err := lost.MarshalBinary()
			require.NoError(t, err)
			require.NoError(t, suite.chain.RedisStorage.Store(
				context.Background(),
				suite.relayerKey(fmt.Sprintf("tx:%d", pending)),
				hexutil.Encode(raw),
				time.Minute,
			))
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			txHash, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			require.NoError(t, err)
			_, _, err = suite.chain.Client.TransactionByHash(context.Background(), lost.Hash())
			require.NoError(t, err)
			tx, _, err := suite.chain.Client.TransactionByHash(context.Background(), txHash)
			require.NoError(t, err)
			assert.Equal(t, pending+1, tx.Nonce())
		})
	})

	t.Run("Given the payer balance is below the amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the transfer rejected", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			payment := suite.chain.FundedPayment(t, 10).BuildPayment()

			_, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			assert.ErrorIs(t, err, apperrors.ErrTransferRejected)
		})

		t.Run("Should leave the relayer nonce unused", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			payment := suite.chain.FundedPayment(t, 10).BuildPayment()
			before, err := suite.chain.Client.PendingNonceAt(context.Background(), suite.chain.Relayer.Address)
			require.NoError(t, err)

			_, _ = suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			after, err := suite.chain.Client.PendingNonceAt(context.Background(), suite.chain.Relayer.Address)
			require.NoError(t, err)
			assert.Equal(t, before, after)
		})
	})

	t.Run("Given the gas budget cannot cover the transfer", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBroadcastingATransferTestingSuite) {
			suite.mockGasBudget.EXPECT().
				Reserve(gomock.Any(), suite.chain.ChainId, gomock.Any(), gomock.Any()).
				Return(apperrors.ErrGasBudgetSpent)
		}

		t.Run("Should return the budget spent error", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			initSuite(suite)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			_, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			assert.ErrorIs(t, err, apperrors.ErrGasBudgetSpent)
		})

		t.Run("Should leave the relayer nonce unused", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			initSuite(suite)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()
			before, err := suite.chain.Client.PendingNonceAt(context.Background(), suite.chain.Relayer.Address)
			require.NoError(t, err)

			_, _ = suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			after, err := suite.chain.Client.PendingNonceAt(context.Background(), suite.chain.Relayer.Address)
			require.NoError(t, err)
			assert.Equal(t, before, after)
		})

		t.Run("Should keep no transaction to resend later", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			initSuite(suite)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()
			before, err := suite.chain.Client.PendingNonceAt(context.Background(), suite.chain.Relayer.Address)
			require.NoError(t, err)

			_, _ = suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			kept, err := suite.chain.RedisStorage.ReadString(
				context.Background(),
				suite.relayerKey(fmt.Sprintf("tx:%d", before)),
			)
			require.NoError(t, err)
			assert.Nil(t, kept)
		})
	})

	t.Run("Given the node refuses the transfer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep no transaction to resend later", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			suite.expectReserved()
			suite.mockGasBudget.EXPECT().Settle(gomock.Any(), suite.chain.ChainId, gomock.Any(), new(big.Int))
			require.NoError(t, suite.chain.Client.Client().Call(
				nil,
				"anvil_setBalance",
				suite.chain.Relayer.Address,
				"0x0",
			))
			pending, err := suite.chain.Client.PendingNonceAt(context.Background(), suite.chain.Relayer.Address)
			require.NoError(t, err)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			_, err = suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			require.ErrorContains(t, err, "send settlement")
			kept, err := suite.chain.RedisStorage.ReadString(
				context.Background(),
				suite.relayerKey(fmt.Sprintf("tx:%d", pending)),
			)
			require.NoError(t, err)
			assert.Nil(t, kept)
		})

		t.Run("Should give its whole gas reservation back", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			suite.expectReserved()
			suite.mockGasBudget.EXPECT().Settle(gomock.Any(), suite.chain.ChainId, gomock.Any(), new(big.Int))
			require.NoError(t, suite.chain.Client.Client().Call(
				nil,
				"anvil_setBalance",
				suite.chain.Relayer.Address,
				"0x0",
			))
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			_, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			require.ErrorContains(t, err, "send settlement")
		})
	})

	t.Run("Given a funded EIP-3009 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should send the transfer", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			suite.expectReserved()
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()

			txHash, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			require.NoError(t, err)
			_, _, err = suite.chain.Client.TransactionByHash(context.Background(), txHash)
			assert.NoError(t, err)
		})

		t.Run("Should reserve gas limit times max fee under the transaction hash", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			payment := suite.chain.FundedPayment(t, 1000).BuildPayment()
			var reservedHash common.Hash
			var reservedCost *big.Int
			suite.mockGasBudget.EXPECT().
				Reserve(gomock.Any(), suite.chain.ChainId, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ uint64, txHash common.Hash, maxCost *big.Int) error {
					reservedHash, reservedCost = txHash, maxCost
					return nil
				})

			txHash, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, payment))

			require.NoError(t, err)
			tx, _, err := suite.chain.Client.TransactionByHash(context.Background(), txHash)
			require.NoError(t, err)
			assert.Equal(t, txHash, reservedHash)
			assert.Equal(t, new(big.Int).Mul(new(big.Int).SetUint64(tx.Gas()), tx.GasFeeCap()), reservedCost)
		})

		t.Run("Should give the next transfer the next relayer nonce", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			suite.expectReserved()
			suite.expectReserved()
			first := suite.chain.FundedPayment(t, 1000).BuildPayment()
			second := suite.chain.FundedPayment(t, 1000).BuildPayment()

			firstHash, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, first))
			require.NoError(t, err)
			secondHash, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId, suite.chain.TransferCall(t, second))

			require.NoError(t, err)
			firstTx, _, err := suite.chain.Client.TransactionByHash(context.Background(), firstHash)
			require.NoError(t, err)
			secondTx, _, err := suite.chain.Client.TransactionByHash(context.Background(), secondHash)
			require.NoError(t, err)
			assert.Equal(t, firstTx.Nonce()+1, secondTx.Nonce())
		})
	})

	t.Run("Given a funded Permit2 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pay the payee through Permit2", func(t *testing.T) {
			t.Parallel()

			suite := WhenBroadcastingATransferBeforeEach(t, nil)
			suite.expectReserved()
			payment := suite.chain.FundedPermit2Payment(t, 1000).BuildPayment()

			_, err := suite.sut.Broadcast(
				context.Background(),
				suite.chain.ChainId,
				suite.chain.TransferCall(t, payment),
			)

			require.NoError(t, err)
			balance, err := suite.chain.Token.BalanceOf(payment.PayTo)
			require.NoError(t, err)
			assert.Equal(t, int64(1000), balance.Int64())
		})
	})
}
