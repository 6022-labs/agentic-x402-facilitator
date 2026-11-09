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

// goldenPermit2Digest comes from ethers.js v6 TypedDataEncoder.hash over the same inputs.
const goldenPermit2Digest = "cc75e64b59b0d9cdaacde47ec4804d1fae01189e6e495f6776b453f85133581a"

func TestWhenHashingAPermitWitnessTransferFrom(t *testing.T) {
	t.Parallel()

	t.Run("Given the reference Permit2 authorization", func(t *testing.T) {
		t.Parallel()

		authorization := &models.TransferAuthorization{
			From:        common.HexToAddress("0x1111111111111111111111111111111111111111"),
			To:          common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
			Value:       big.NewInt(10000),
			ValidAfter:  big.NewInt(1700000000),
			ValidBefore: big.NewInt(1700000300),
		}
		big.NewInt(42).FillBytes(authorization.Nonce[:])

		t.Run("Should hash to the digest ethers.js produces", func(t *testing.T) {
			t.Parallel()

			digest, err := helpers.PermitWitnessTransferFromDigest(
				authorization,
				common.HexToAddress("0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"),
				80002,
			)

			assert.NoError(t, err)
			assert.Equal(t, goldenPermit2Digest, hex.EncodeToString(digest))
		})
	})
}
