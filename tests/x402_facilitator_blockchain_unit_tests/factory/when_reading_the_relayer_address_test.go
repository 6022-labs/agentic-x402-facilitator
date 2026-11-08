package factory_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReadingTheRelayerAddressTestingSuite struct {
	mockRelayerKeyStorage *interfaces_mocks.MockRelayerKeyStorageInterface
	sut                   *factory.TransactorFactory
}

func WhenReadingTheRelayerAddressBeforeEach(t *testing.T) *WhenReadingTheRelayerAddressTestingSuite {
	mockRelayerKeyStorage := interfaces_mocks.NewMockRelayerKeyStorageInterface(gomock.NewController(t))
	return &WhenReadingTheRelayerAddressTestingSuite{
		mockRelayerKeyStorage: mockRelayerKeyStorage,
		sut:                   factory.NewTransactorFactory(mockRelayerKeyStorage, zap.NewNop()),
	}
}

func TestWhenReadingTheRelayerAddress(t *testing.T) {
	t.Parallel()

	t.Run("Given the relayer key failed to load moments ago", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReadingTheRelayerAddressTestingSuite) {
			suite.mockRelayerKeyStorage.EXPECT().
				ReadPrivateKey(gomock.Any()).
				Return("", errors.New("vault sealed")).
				Times(1)
			_, _ = suite.sut.Address(context.Background())
		}

		t.Run("Should fail without reading the key again", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerAddressBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Address(context.Background())

			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
		})
	})

	t.Run("Given the relayer key cannot be read", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReadingTheRelayerAddressTestingSuite) {
			suite.mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("", errors.New("vault sealed"))
		}

		t.Run("Should return a relayer unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerAddressBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Address(context.Background())

			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
		})
	})

	t.Run("Given the stored relayer key is not a private key", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReadingTheRelayerAddressTestingSuite) {
			suite.mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return("0xnothex", nil)
		}

		t.Run("Should return a relayer unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerAddressBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Address(context.Background())

			assert.ErrorIs(t, err, apperrors.ErrRelayerUnavailable)
		})
	})

	t.Run("Given the relayer key is stored", func(t *testing.T) {
		t.Parallel()

		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		initSuite := func(suite *WhenReadingTheRelayerAddressTestingSuite) {
			suite.mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).
				Return(hexutil.Encode(crypto.FromECDSA(key)), nil)
		}

		t.Run("Should return the key's address", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerAddressBeforeEach(t)
			initSuite(suite)

			address, err := suite.sut.Address(context.Background())

			require.NoError(t, err)
			assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey), address)
		})
	})
}
