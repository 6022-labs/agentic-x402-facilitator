package services

import (
	"context"
	"math/big"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"go.uber.org/zap"
)

// RelayerBalanceReader logs and skips a chain it cannot read: it only feeds a gauge.
type RelayerBalanceReader struct {
	ethClientFactory  *factory.EthClientFactory
	transactorFactory *factory.TransactorFactory
	chainsSettings    *blockchain_settings.ChainsSettings
	logger            *zap.Logger
}

func NewRelayerBalanceReader(
	ethClientFactory *factory.EthClientFactory,
	transactorFactory *factory.TransactorFactory,
	chainsSettings *blockchain_settings.ChainsSettings,
	logger *zap.Logger,
) *RelayerBalanceReader {
	return &RelayerBalanceReader{
		ethClientFactory:  ethClientFactory,
		transactorFactory: transactorFactory,
		chainsSettings:    chainsSettings,
		logger:            logger,
	}
}

func (r *RelayerBalanceReader) Balances(ctx context.Context) map[uint64]*big.Int {
	balances := map[uint64]*big.Int{}
	relayer, err := r.transactorFactory.Address(ctx)
	if err != nil {
		r.logger.Warn("relayer balance not read: relayer unavailable", zap.Error(err))
		return balances
	}

	for _, chainId := range r.chainsSettings.ChainIds() {
		client, err := r.ethClientFactory.For(chainId)
		if err != nil {
			r.logger.Warn("relayer balance not read", zap.Uint64("chainId", chainId), zap.Error(err))
			continue
		}
		readCtx, cancel := context.WithTimeout(ctx, r.chainsSettings.Get(chainId).RequestTimeout)
		balance, err := client.BalanceAt(readCtx, relayer, nil)
		cancel()
		if err != nil {
			r.logger.Warn("relayer balance not read", zap.Uint64("chainId", chainId), zap.Error(err))
			continue
		}
		balances[chainId] = balance
	}
	return balances
}
