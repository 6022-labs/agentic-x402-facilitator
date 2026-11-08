package factory_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenCreatingATransactorTestingSuite struct {
	mockRelayerKeyStorage *interfaces_mocks.MockRelayerKeyStorageInterface
	sut                   *factory.TransactorFactory
}

func WhenCreatingATransactorBeforeEach(t *testing.T) *WhenCreatingATransactorTestingSuite {
	mockRelayerKeyStorage := interfaces_mocks.NewMockRelayerKeyStorageInterface(gomock.NewController(t))
	return &WhenCreatingATransactorTestingSuite{
		mockRelayerKeyStorage: mockRelayerKeyStorage,
		sut:                   factory.NewTransactorFactory(mockRelayerKeyStorage, zap.NewNop()),
	}
}

func TestWhenCreatingATransactor(t *testing.T) {
	t.Parallel()

	t.Run("Given the relayer key cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("vault sealed")

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("", sourceErr)
		}

		t.Run("Should return a relayer unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			opts, err := suite.sut.Create(context.Background(), 80002)

			assert.Nil(t, opts)
			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
			assert.ErrorIs(t, err, sourceErr)
		})

		t.Run("Should fail the next call at once without reading it again", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			_, _ = suite.sut.Create(context.Background(), 80002)
			_, err := suite.sut.Create(context.Background(), 80002)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the stored relayer key is not a private key", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("0xnothex", nil)
		}

		t.Run("Should return a relayer unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), 80002)

			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
		})
	})

	t.Run("Given a 0x-prefixed relayer key with surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		key, err := crypto.GenerateKey()
		require.NoError(t, err)
		storedKey := "  " + hexutil.Encode(crypto.FromECDSA(key)) + "  "

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return(storedKey, nil).Times(1)
		}

		t.Run("Should sign as the relayer for the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			opts, err := suite.sut.Create(context.Background(), 80002)

			require.NoError(t, err)
			assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey), opts.From)
			signed, err := opts.Signer(opts.From, types.NewTx(&types.DynamicFeeTx{ChainID: big.NewInt(80002)}))
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(80002), signed.ChainId())
		})

		t.Run("Should read the key only once", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			_, _ = suite.sut.Create(context.Background(), 80002)
			_, err := suite.sut.Create(context.Background(), 80002)

			assert.NoError(t, err)
		})
	})
}
