package helpers

import (
	"fmt"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// TransferWithAuthorizationDigest hashes the EIP-3009 typed data under the asset's EIP-712 domain.
func TransferWithAuthorizationDigest(
	authorization *models.TransferAuthorization,
	asset common.Address,
	eip712Name string,
	eip712Version string,
	chainId uint64,
) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"TransferWithAuthorization": {
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "validAfter", Type: "uint256"},
				{Name: "validBefore", Type: "uint256"},
				{Name: "nonce", Type: "bytes32"},
			},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain: apitypes.TypedDataDomain{
			Name:              eip712Name,
			Version:           eip712Version,
			ChainId:           math.NewHexOrDecimal256(int64(chainId)),
			VerifyingContract: asset.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"from":        authorization.From.Hex(),
			"to":          authorization.To.Hex(),
			"value":       authorization.Value.String(),
			"validAfter":  authorization.ValidAfter.String(),
			"validBefore": authorization.ValidBefore.String(),
			"nonce":       hexutil.Encode(authorization.Nonce[:]),
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash typed data: %w", err)
	}
	return digest, nil
}
