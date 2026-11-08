package responses

import (
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

// signersEvmNetworks is the CAIP-2 pattern covering every EVM chain: the relayer signs with one key everywhere.
const signersEvmNetworks = "eip155:*"

type SupportedKind struct {
	X402Version int    `json:"x402Version"`
	Scheme      string `json:"scheme"`
	Network     string `json:"network"`
}

// GetSupportedResponse is the x402 v2 SupportedResponse.
type GetSupportedResponse struct {
	Kinds      []SupportedKind     `json:"kinds"`
	Extensions []string            `json:"extensions"`
	Signers    map[string][]string `json:"signers"`
}

func NewGetSupportedResponse(chainIds []uint64, relayer common.Address) *GetSupportedResponse {
	return &GetSupportedResponse{
		Kinds: lo.Map(chainIds, func(chainId uint64, _ int) SupportedKind {
			return SupportedKind{
				X402Version: enums.X402Version,
				Scheme:      enums.PaymentSchemeExact.String(),
				Network:     helpers.FormatCaip2(chainId),
			}
		}),
		Extensions: []string{},
		Signers:    map[string][]string{signersEvmNetworks: {relayer.Hex()}},
	}
}
