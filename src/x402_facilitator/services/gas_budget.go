package services

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const gasBudgetWindow = 24 * time.Hour

var weiPerGwei = big.NewInt(1_000_000_000)

// GasBudget counts gwei; a reservation it cannot correct stays at its worst case until the day ends.
type GasBudget struct {
	gasBudgetSettings *x402_facilitator_settings.GasBudgetSettings
	counterStorage    x402_facilitator_interfaces.CounterStorageInterface
	cacheStorage      x402_facilitator_interfaces.CacheStorageInterface
	logger            *zap.Logger
	now               func() time.Time
}

func NewGasBudget(
	gasBudgetSettings *x402_facilitator_settings.GasBudgetSettings,
	counterStorage x402_facilitator_interfaces.CounterStorageInterface,
	cacheStorage x402_facilitator_interfaces.CacheStorageInterface,
	logger *zap.Logger,
) *GasBudget {
	return &GasBudget{
		gasBudgetSettings: gasBudgetSettings,
		counterStorage:    counterStorage,
		cacheStorage:      cacheStorage,
		logger:            logger,
		now:               time.Now,
	}
}

func (b *GasBudget) Check(ctx context.Context, chainId uint64) error {
	budget, bounded := b.gasBudgetSettings.Get(chainId)
	if !bounded {
		return nil
	}
	spent, err := b.counterStorage.Read(ctx, b.dayKey(chainId))
	if err != nil {
		return apperrors.NewUnavailableError(
			"store_unavailable",
			"the settlement store is unavailable",
			fmt.Errorf("gas budget: %w", err),
		)
	}
	if spent >= budget {
		return apperrors.NewUnavailableError(
			"gas_budget_spent",
			"the chain's daily gas budget is spent, retry tomorrow",
			apperrors.ErrGasBudgetSpent,
		)
	}
	return nil
}

func (b *GasBudget) Reserve(ctx context.Context, chainId uint64, txHash common.Hash, maxCost *big.Int) error {
	budget, bounded := b.gasBudgetSettings.Get(chainId)
	if !bounded {
		return nil
	}
	amount := b.gwei(maxCost)
	if !amount.IsInt64() || amount.Int64() > budget {
		return apperrors.ErrGasBudgetSpent
	}
	dayKey := b.dayKey(chainId)
	reserved, err := b.counterStorage.TryAdd(ctx, dayKey, amount.Int64(), budget, gasBudgetWindow)
	if err != nil {
		return fmt.Errorf("reserve gas budget: %w", err)
	}
	if !reserved {
		return apperrors.ErrGasBudgetSpent
	}
	// Lost, only the correction is: the reservation then counts at its worst case.
	reservation := dayKey + " " + amount.String()
	if err := b.cacheStorage.Store(ctx, b.reservationKey(chainId, txHash), reservation, gasBudgetWindow); err != nil {
		b.logger.Warn("failed to keep a gas reservation", zap.String("tx", txHash.Hex()), zap.Error(err))
	}
	return nil
}

func (b *GasBudget) Settle(ctx context.Context, chainId uint64, txHash common.Hash, cost *big.Int) {
	if _, bounded := b.gasBudgetSettings.Get(chainId); !bounded {
		return
	}
	if err := b.settle(ctx, chainId, txHash, b.gwei(cost)); err != nil {
		b.logger.Warn("failed to correct a gas reservation", zap.String("tx", txHash.Hex()), zap.Error(err))
	}
}

func (b *GasBudget) settle(ctx context.Context, chainId uint64, txHash common.Hash, cost *big.Int) error {
	// Taken first and atomically: a failure after it leaves the worst case counted, never a double correction.
	stored, err := b.cacheStorage.ReadAndDelete(ctx, b.reservationKey(chainId, txHash))
	if err != nil || stored == nil {
		return err
	}
	dayKey, amount, _ := strings.Cut(*stored, " ")
	reserved, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return fmt.Errorf("read gas reservation: %w", err)
	}
	if !cost.IsInt64() || cost.Int64() >= reserved {
		return nil
	}
	return b.counterStorage.Subtract(ctx, dayKey, reserved-cost.Int64())
}

// gwei rounds up, so many tiny costs never count as free.
func (b *GasBudget) gwei(wei *big.Int) *big.Int {
	quotient, remainder := new(big.Int).QuoRem(wei, weiPerGwei, new(big.Int))
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func (b *GasBudget) dayKey(chainId uint64) string {
	return fmt.Sprintf("x402:gas:%d:%d", chainId, b.now().Unix()/int64(gasBudgetWindow/time.Second))
}

func (b *GasBudget) reservationKey(chainId uint64, txHash common.Hash) string {
	return fmt.Sprintf("x402:gas:%d:tx:%s", chainId, txHash.Hex())
}
