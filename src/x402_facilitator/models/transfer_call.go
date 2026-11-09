package models

import "github.com/ethereum/go-ethereum/common"

// TransferCall is the transaction the relayer sends to settle a payment.
type TransferCall struct {
	To   common.Address
	Data []byte
}
