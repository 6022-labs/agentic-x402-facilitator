package models

import (
	"fmt"
	"math/big"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/ethereum/go-ethereum/common"
)

// Payment is a structurally valid x402 exact payment, authorized through Method.
type Payment struct {
	ChainId       uint64
	Asset         common.Address
	PayTo         common.Address
	Amount        *big.Int
	Method        enums.AssetTransferMethod
	Authorization *TransferAuthorization
}

// SettlementKey identifies the authorization on-chain: EIP-3009 consumes a nonce per token, Permit2 per owner.
func (p *Payment) SettlementKey() string {
	scope := p.Asset.Hex()
	if p.Method == enums.AssetTransferMethodPermit2 {
		scope = "permit2"
	}
	return fmt.Sprintf(
		"%d:%s:%s:%s",
		p.ChainId,
		scope,
		p.Authorization.From.Hex(),
		common.Bytes2Hex(p.Authorization.Nonce[:]),
	)
}

// Terms is what the payer signed besides its nonce, token included since a Permit2 nonce spans tokens.
func (p *Payment) Terms() string {
	return fmt.Sprintf("%s:%s:%s", p.Method, p.Asset.Hex(), p.Authorization.Terms())
}
