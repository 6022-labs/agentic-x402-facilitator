package helpers

import (
	"fmt"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	// Permit2Address is Uniswap's canonical Permit2, at the same address on every chain.
	Permit2Address = "0x000000000022D473030F116dDEE9F6B43aC78BA3"
	// X402ExactPermit2ProxyAddress is the only spender a payer may sign for: it sends to the witness's to.
	X402ExactPermit2ProxyAddress = "0x402085c248EeA27D92E8b30b2C58ed07f9E20001"
)

// PermitWitnessTransferFromDigest hashes the x402 Permit2 typed data; Permit2's domain has no version field.
func PermitWitnessTransferFromDigest(
	authorization *models.TransferAuthorization,
	token common.Address,
	chainId uint64,
) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"PermitWitnessTransferFrom": {
				{Name: "permitted", Type: "TokenPermissions"},
				{Name: "spender", Type: "address"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
				{Name: "witness", Type: "Witness"},
			},
			"TokenPermissions": {
				{Name: "token", Type: "address"},
				{Name: "amount", Type: "uint256"},
			},
			"Witness": {
				{Name: "to", Type: "address"},
				{Name: "validAfter", Type: "uint256"},
			},
		},
		PrimaryType: "PermitWitnessTransferFrom",
		Domain: apitypes.TypedDataDomain{
			Name:              "Permit2",
			ChainId:           math.NewHexOrDecimal256(int64(chainId)),
			VerifyingContract: Permit2Address,
		},
		Message: apitypes.TypedDataMessage{
			"permitted": apitypes.TypedDataMessage{
				"token":  token.Hex(),
				"amount": authorization.Value.String(),
			},
			"spender":  X402ExactPermit2ProxyAddress,
			"nonce":    authorization.NonceInt().String(),
			"deadline": authorization.ValidBefore.String(),
			"witness": apitypes.TypedDataMessage{
				"to":         authorization.To.Hex(),
				"validAfter": authorization.ValidAfter.String(),
			},
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash permit2 typed data: %w", err)
	}
	return digest, nil
}
