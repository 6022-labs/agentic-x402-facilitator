package models

import "github.com/ethereum/go-ethereum/common"

type TransferStatus string

const (
	TransferStatusConfirmed TransferStatus = "confirmed"
	TransferStatusReverted  TransferStatus = "reverted"
	// TransferStatusPending is a broadcast transfer still unconfirmed when the receipt wait ended.
	TransferStatusPending TransferStatus = "pending"
)

// Transfer is a broadcast transferWithAuthorization and how far it got.
type Transfer struct {
	TxHash common.Hash
	Status TransferStatus
}
