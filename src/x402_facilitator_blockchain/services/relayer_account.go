package services

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	"github.com/ethereum/go-ethereum/common"
)

type RelayerAccount struct {
	transactorFactory *factory.TransactorFactory
}

func NewRelayerAccount(transactorFactory *factory.TransactorFactory) *RelayerAccount {
	return &RelayerAccount{transactorFactory: transactorFactory}
}

func (a *RelayerAccount) Address(ctx context.Context) (common.Address, error) {
	return a.transactorFactory.Address(ctx)
}
