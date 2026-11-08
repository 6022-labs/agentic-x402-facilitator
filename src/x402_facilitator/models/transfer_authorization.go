package models

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// TransferAuthorization is a parsed EIP-3009 TransferWithAuthorization with its 65-byte signature.
type TransferAuthorization struct {
	From        common.Address
	To          common.Address
	Value       *big.Int
	ValidAfter  *big.Int
	ValidBefore *big.Int
	Nonce       [32]byte
	Signature   []byte
}

// Terms is what the payer signed besides (from, nonce): a payer can sign other terms under a used nonce.
func (a *TransferAuthorization) Terms() string {
	return fmt.Sprintf("%s:%s:%s:%s", a.To.Hex(), a.Value, a.ValidAfter, a.ValidBefore)
}

// neverExpires stays far below where time.Unix overflows.
const neverExpires = math.MaxInt64 / 2

// ExpiresAt is validBefore as a time; a value past neverExpires is read as never expiring.
func (a *TransferAuthorization) ExpiresAt() time.Time {
	if !a.ValidBefore.IsInt64() || a.ValidBefore.Int64() > neverExpires {
		return time.Unix(neverExpires, 0)
	}
	return time.Unix(a.ValidBefore.Int64(), 0)
}
