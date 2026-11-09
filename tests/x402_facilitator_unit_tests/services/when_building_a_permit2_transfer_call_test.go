package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenBuildingAPermit2TransferCallTestingSuite struct {
	sut *x402_facilitator_services.Permit2AssetTransferMethodStrategy
}

func WhenBuildingAPermit2TransferCallBeforeEach() *WhenBuildingAPermit2TransferCallTestingSuite {
	return &WhenBuildingAPermit2TransferCallTestingSuite{
		sut: x402_facilitator_services.NewPermit2AssetTransferMethodStrategy(zap.NewNop(), nil, nil),
	}
}

func TestWhenBuildingAPermit2TransferCall(t *testing.T) {
	t.Parallel()

	selector := crypto.Keccak256([]byte("settle(((address,uint256),uint256,uint256),address,(address,uint256),bytes)"))[:4]

	t.Run("Given a Permit2 payment signed with a v of 0 or 1", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithPermit2().WithRawRecoveryId().BuildPayment()

		t.Run("Should hand Permit2 a v of 27 or 28", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAPermit2TransferCallBeforeEach()

			call, err := suite.sut.TransferCall(payment)

			require.NoError(t, err)
			// The 65-byte signature ends the calldata, padded to 96 bytes after its length word.
			v := call.Data[len(call.Data)-32]
			assert.Contains(t, []byte{27, 28}, v)
		})
	})

	t.Run("Given a Permit2 payment", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().WithPermit2().BuildPayment()

		t.Run("Should call settle on x402ExactPermit2Proxy", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAPermit2TransferCallBeforeEach()

			call, err := suite.sut.TransferCall(payment)

			require.NoError(t, err)
			assert.Equal(t, common.HexToAddress(helpers.X402ExactPermit2ProxyAddress), call.To)
			assert.Equal(t, selector, call.Data[:4])
		})
	})
}
