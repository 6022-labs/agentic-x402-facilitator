package helpers_test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// goldenEip3009Digest comes from ethers.js v6 TypedDataEncoder.hash over the same inputs.
const goldenEip3009Digest = "c08713f75413a1225021aaf18ddb94e18ef4a82b9208f345e773a7042cc74b82"

func TestWhenHashingATransferWithAuthorization(t *testing.T) {
	t.Parallel()

	t.Run("Given the reference EIP-3009 authorization", func(t *testing.T) {
		t.Parallel()

		authorization := &models.TransferAuthorization{
			From:        common.HexToAddress("0x1111111111111111111111111111111111111111"),
			To:          common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
			Value:       big.NewInt(10000),
			ValidAfter:  big.NewInt(1700000000),
			ValidBefore: big.NewInt(1700000300),
		}

		t.Run("Should hash to the digest ethers.js produces", func(t *testing.T) {
			t.Parallel()

			digest, err := helpers.TransferWithAuthorizationDigest(
				authorization,
				common.HexToAddress("0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"),
				"USDC",
				"2",
				80002,
			)

			assert.NoError(t, err)
			assert.Equal(t, goldenEip3009Digest, hex.EncodeToString(digest))
		})
	})
}
