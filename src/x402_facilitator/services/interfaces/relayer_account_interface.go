package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// RelayerAccountInterface is the wallet that submits settlements and pays their gas.
type RelayerAccountInterface interface {
	Address(ctx context.Context) (common.Address, error)
}
