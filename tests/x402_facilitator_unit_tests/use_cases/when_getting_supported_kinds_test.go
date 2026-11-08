package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/responses"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenGettingSupportedKindsTestingSuite struct {
	mockRelayerAccount *interfaces_mocks.MockRelayerAccountInterface
	sut                *use_cases.GetSupported
}

func WhenGettingSupportedKindsBeforeEach(t *testing.T) *WhenGettingSupportedKindsTestingSuite {
	mockRelayerAccount := interfaces_mocks.NewMockRelayerAccountInterface(gomock.NewController(t))
	tokensSettings := &x402_facilitator_settings.TokensSettings{
		80002: {common_tests.RandomEthAddress(): {Eip712Name: "USDC", Eip712Version: "2"}},
		137:   {common_tests.RandomEthAddress(): {Eip712Name: "USD Coin", Eip712Version: "2"}},
	}

	return &WhenGettingSupportedKindsTestingSuite{
		mockRelayerAccount: mockRelayerAccount,
		sut:                use_cases.NewGetSupported(tokensSettings, mockRelayerAccount),
	}
}

func TestWhenGettingSupportedKinds(t *testing.T) {
	t.Parallel()

	t.Run("Given the relayer wallet cannot be loaded", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("vault sealed")

		initSuite := func(suite *WhenGettingSupportedKindsTestingSuite) {
			suite.mockRelayerAccount.EXPECT().Address(gomock.Any()).Return(common.Address{}, sourceErr)
		}

		t.Run("Should return relayer_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingSupportedKindsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.Nil(t, response)
			var unavailableErr *apperrors.UnavailableError
			require.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "relayer_unavailable", unavailableErr.Code)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the relayer wallet is loaded", func(t *testing.T) {
		t.Parallel()

		relayer := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenGettingSupportedKindsTestingSuite) {
			suite.mockRelayerAccount.EXPECT().Address(gomock.Any()).Return(relayer, nil)
		}

		t.Run("Should list one exact kind per network", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingSupportedKindsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			require.NoError(t, err)
			assert.Equal(t, []responses.SupportedKind{
				{X402Version: 2, Scheme: "exact", Network: "eip155:137"},
				{X402Version: 2, Scheme: "exact", Network: "eip155:80002"},
			}, response.Kinds)
		})

		t.Run("Should list no extension", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingSupportedKindsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			require.NoError(t, err)
			assert.Empty(t, response.Extensions)
		})

		t.Run("Should name the relayer as signer", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingSupportedKindsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			require.NoError(t, err)
			assert.Equal(t, map[string][]string{"eip155:*": {relayer.Hex()}}, response.Signers)
		})
	})
}
