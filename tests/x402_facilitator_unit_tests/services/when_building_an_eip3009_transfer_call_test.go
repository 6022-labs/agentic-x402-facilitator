package services_test

import (
	"testing"

	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenBuildingAnEip3009TransferCallTestingSuite struct {
	sut *x402_facilitator_services.Eip3009AssetTransferMethodStrategy
}

func WhenBuildingAnEip3009TransferCallBeforeEach() *WhenBuildingAnEip3009TransferCallTestingSuite {
	return &WhenBuildingAnEip3009TransferCallTestingSuite{
		sut: x402_facilitator_services.NewEip3009AssetTransferMethodStrategy(zap.NewNop(), nil),
	}
}

func TestWhenBuildingAnEip3009TransferCall(t *testing.T) {
	t.Parallel()

	t.Run("Given an EIP-3009 payment", func(t *testing.T) {
		t.Parallel()

		payment := model_builders.NewPaymentRequestBuilder().BuildPayment()
		selector := crypto.Keccak256(
			[]byte("transferWithAuthorization(address,address,uint256,uint256,uint256,bytes32,uint8,bytes32,bytes32)"),
		)[:4]

		t.Run("Should call transferWithAuthorization on the token", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAnEip3009TransferCallBeforeEach()

			call, err := suite.sut.TransferCall(payment)

			require.NoError(t, err)
			assert.Equal(t, payment.Asset, call.To)
			assert.Equal(t, selector, call.Data[:4])
		})
	})
}
