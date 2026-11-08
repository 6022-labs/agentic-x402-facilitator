package transports_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/transports"
	metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenRoundTrippingAnRpcRequestTestingSuite struct {
	mockBlockchainRpcMetrics *metrics_mocks.MockBlockchainRpcMetricsInterface
}

func WhenRoundTrippingAnRpcRequestBeforeEach(t *testing.T) *WhenRoundTrippingAnRpcRequestTestingSuite {
	return &WhenRoundTrippingAnRpcRequestTestingSuite{
		mockBlockchainRpcMetrics: metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t)),
	}
}

// serve starts an RPC stand-in that hands each request it receives to received.
func (suite *WhenRoundTrippingAnRpcRequestTestingSuite) serve(t *testing.T, received **http.Request) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*received = r
	}))
	t.Cleanup(server.Close)
	return server
}

func (suite *WhenRoundTrippingAnRpcRequestTestingSuite) send(t *testing.T, rpcUrl string) {
	sut, err := transports.NewRpcRoundTripper(http.DefaultTransport, rpcUrl, suite.mockBlockchainRpcMetrics, "80002")
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, sut.Origin(), strings.NewReader("{}"))
	require.NoError(t, err)
	response, err := sut.RoundTrip(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
}

func (suite *WhenRoundTrippingAnRpcRequestTestingSuite) expectAnyRecord() {
	suite.mockBlockchainRpcMetrics.EXPECT().
		RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
}

func TestWhenRoundTrippingAnRpcRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given an RPC URL with an API key in its path and query", func(t *testing.T) {
		t.Parallel()

		t.Run("Should send to the full URL", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnRpcRequestBeforeEach(t)
			var received *http.Request
			server := suite.serve(t, &received)
			suite.expectAnyRecord()

			suite.send(t, server.URL+"/v2/secret?key=other")

			assert.Equal(t, "/v2/secret", received.URL.Path)
			assert.Equal(t, "key=other", received.URL.RawQuery)
		})

		t.Run("Should record the call under the origin only", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnRpcRequestBeforeEach(t)
			var received *http.Request
			server := suite.serve(t, &received)
			suite.mockBlockchainRpcMetrics.EXPECT().
				RecordRequest(gomock.Any(), "80002", server.URL, gomock.Any(), http.StatusOK, nil)

			suite.send(t, server.URL+"/v2/secret?key=other")
		})
	})

	t.Run("Given an RPC URL with credentials", func(t *testing.T) {
		t.Parallel()

		t.Run("Should send them as basic auth", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnRpcRequestBeforeEach(t)
			var received *http.Request
			server := suite.serve(t, &received)
			suite.expectAnyRecord()

			suite.send(t, strings.Replace(server.URL, "://", "://user:pass@", 1))

			username, password, ok := received.BasicAuth()
			require.True(t, ok)
			assert.Equal(t, "user", username)
			assert.Equal(t, "pass", password)
		})
	})
}
