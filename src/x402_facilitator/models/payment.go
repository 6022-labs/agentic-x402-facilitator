package models

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Payment is a structurally valid x402 exact EIP-3009 payment.
type Payment struct {
	ChainId       uint64
	Asset         common.Address
	PayTo         common.Address
	Amount        *big.Int
	Authorization *TransferAuthorization
}

// SettlementKey identifies the authorization on-chain: a token consumes each (from, nonce) once.
func (p *Payment) SettlementKey() string {
	return fmt.Sprintf(
		"%d:%s:%s:%s",
		p.ChainId,
		p.Asset.Hex(),
		p.Authorization.From.Hex(),
		common.Bytes2Hex(p.Authorization.Nonce[:]),
	)
}
