package services

import (
	"context"
	"strings"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// maxRemember caps how long a settlement's transaction is remembered for retries.
const maxRemember = time.Hour

// SettlementGuard logs a failed release instead of returning it: the lock expires anyway.
type SettlementGuard struct {
	lockStorage        x402_facilitator_interfaces.LockStorageInterface
	cacheStorage       x402_facilitator_interfaces.CacheStorageInterface
	settlementSettings *x402_facilitator_settings.SettlementSettings
	logger             *zap.Logger
}

func NewSettlementGuard(
	lockStorage x402_facilitator_interfaces.LockStorageInterface,
	cacheStorage x402_facilitator_interfaces.CacheStorageInterface,
	settlementSettings *x402_facilitator_settings.SettlementSettings,
	logger *zap.Logger,
) *SettlementGuard {
	return &SettlementGuard{
		lockStorage:        lockStorage,
		cacheStorage:       cacheStorage,
		settlementSettings: settlementSettings,
		logger:             logger,
	}
}

func (g *SettlementGuard) TryAcquire(ctx context.Context, key string) (string, bool, error) {
	return g.lockStorage.TryLock(ctx, g.lockKey(key), g.settlementSettings.LockTtl)
}

func (g *SettlementGuard) Release(ctx context.Context, key string, token string) {
	if err := g.lockStorage.Unlock(context.WithoutCancel(ctx), g.lockKey(key), token); err != nil {
		g.logger.Warn("failed to release a settlement lock", zap.String("key", key), zap.Error(err))
	}
}

func (g *SettlementGuard) RememberTransaction(
	ctx context.Context,
	payment *models.Payment,
	txHash common.Hash,
) error {
	ttl := min(max(time.Until(payment.Authorization.ExpiresAt()), g.settlementSettings.LockTtl), maxRemember)
	return g.cacheStorage.Store(
		context.WithoutCancel(ctx),
		g.transactionKey(payment.SettlementKey()),
		txHash.Hex()+" "+payment.Terms(),
		ttl,
	)
}

func (g *SettlementGuard) RememberedTransaction(ctx context.Context, payment *models.Payment) (*common.Hash, error) {
	stored, err := g.cacheStorage.ReadString(ctx, g.transactionKey(payment.SettlementKey()))
	if err != nil || stored == nil {
		return nil, err
	}
	hash, terms, _ := strings.Cut(*stored, " ")
	if terms != payment.Terms() {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidTransactionState,
			"this nonce was already used by another authorization",
		)
	}
	txHash := common.HexToHash(hash)
	return &txHash, nil
}

func (g *SettlementGuard) lockKey(key string) string {
	return "x402:settlement:" + key
}

func (g *SettlementGuard) transactionKey(key string) string {
	return "x402:settlement:" + key + ":tx"
}
