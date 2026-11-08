package factory_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	blockchain_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenGettingAnEthClientTestingSuite struct {
	sut *factory.EthClientFactory
}

func WhenGettingAnEthClientBeforeEach(
	t *testing.T,
	chainsSettings blockchain_settings.ChainsSettings,
) *WhenGettingAnEthClientTestingSuite {
	mockController := gomock.NewController(t)

	mockBlockchainRpcMetrics := blockchain_mocks.NewMockBlockchainRpcMetricsInterface(mockController)
	mockBlockchainRpcMetrics.EXPECT().
		RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	sut := factory.NewEthClientFactory(&chainsSettings, mockBlockchainRpcMetrics)

	return &WhenGettingAnEthClientTestingSuite{sut: sut}
}

func TestWhenGettingAnEthClient(t *testing.T) {
	t.Parallel()

	t.Run("Given the chain is not configured", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error rather than a nil client", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnEthClientBeforeEach(t, blockchain_settings.ChainsSettings{})

			client, err := suite.sut.For(80002)

			assert.ErrorContains(t, err, "no RPC configured for chain 80002")
			assert.Nil(t, client)
		})
	})

	t.Run("Given an RPC URL that is not a URL", func(t *testing.T) {
		t.Parallel()

		chainsSettings := blockchain_settings.ChainsSettings{
			80002: {RpcHttpUrl: "not a url", RequestTimeout: time.Second},
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnEthClientBeforeEach(t, chainsSettings)

			_, err := suite.sut.For(80002)

			assert.ErrorContains(t, err, "rpc url of chain 80002 is not a URL")
		})
	})

	t.Run("Given an unreachable RPC URL carrying an API key", func(t *testing.T) {
		t.Parallel()

		chainsSettings := blockchain_settings.ChainsSettings{
			80002: {RpcHttpUrl: "http://127.0.0.1:1/v2/secret-api-key", RequestTimeout: time.Second},
		}

		t.Run("Should keep the key out of its errors", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnEthClientBeforeEach(t, chainsSettings)
			client, err := suite.sut.For(80002)
			require.NoError(t, err)

			_, err = client.ChainID(context.Background())

			require.Error(t, err)
			assert.NotContains(t, err.Error(), "secret-api-key")
		})
	})

	t.Run("Given a configured chain", func(t *testing.T) {
		t.Parallel()

		chainsSettings := blockchain_settings.ChainsSettings{
			80002: {RpcHttpUrl: "http://127.0.0.1:1", RequestTimeout: time.Second},
		}

		t.Run("Should dial it once and reuse the client", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnEthClientBeforeEach(t, chainsSettings)

			first, err := suite.sut.For(80002)
			require.NoError(t, err)
			second, err := suite.sut.For(80002)

			require.NoError(t, err)
			assert.Same(t, first, second)
		})
	})
}
