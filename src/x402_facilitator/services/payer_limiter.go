package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"go.uber.org/zap"
)

const payerWindow = time.Hour

// PayerLimiter counts in fixed clock-hour windows; a payer is trusted only once its signature was verified.
type PayerLimiter struct {
	settlementSettings *x402_facilitator_settings.SettlementSettings
	counterStorage     x402_facilitator_interfaces.CounterStorageInterface
	logger             *zap.Logger
	now                func() time.Time
}

func NewPayerLimiter(
	settlementSettings *x402_facilitator_settings.SettlementSettings,
	counterStorage x402_facilitator_interfaces.CounterStorageInterface,
	logger *zap.Logger,
) *PayerLimiter {
	return &PayerLimiter{
		settlementSettings: settlementSettings,
		counterStorage:     counterStorage,
		logger:             logger,
		now:                time.Now,
	}
}

func (l *PayerLimiter) Check(ctx context.Context, payment *models.Payment) error {
	if l.settlementSettings.MaxPerPayerPerHour == 0 {
		return nil
	}
	count, err := l.counterStorage.Read(ctx, l.key(payment))
	if err != nil {
		return l.unavailable(err)
	}
	if count >= l.settlementSettings.MaxPerPayerPerHour {
		return l.reached()
	}
	return nil
}

func (l *PayerLimiter) Take(ctx context.Context, payment *models.Payment) (string, error) {
	if l.settlementSettings.MaxPerPayerPerHour == 0 {
		return "", nil
	}
	key := l.key(payment)
	taken, err := l.counterStorage.TryAdd(ctx, key, 1, l.settlementSettings.MaxPerPayerPerHour, payerWindow)
	if err != nil {
		return "", l.unavailable(err)
	}
	if !taken {
		return "", l.reached()
	}
	return key, nil
}

func (l *PayerLimiter) GiveBack(ctx context.Context, hold string) {
	if hold == "" {
		return
	}
	if err := l.counterStorage.Subtract(context.WithoutCancel(ctx), hold, 1); err != nil {
		l.logger.Warn("failed to uncount a settlement", zap.String("key", hold), zap.Error(err))
	}
}

func (l *PayerLimiter) key(payment *models.Payment) string {
	return fmt.Sprintf(
		"x402:payer:%d:%s:%d",
		payment.ChainId,
		strings.ToLower(payment.Authorization.From.Hex()),
		l.now().Unix()/int64(payerWindow/time.Second),
	)
}

func (l *PayerLimiter) reached() error {
	return apperrors.NewInvalidPaymentError(
		enums.PaymentReasonPayerLimitReached,
		"the payer reached its settlements for this hour, retry later",
	)
}

func (l *PayerLimiter) unavailable(err error) error {
	return apperrors.NewUnavailableError(
		"store_unavailable",
		"the settlement store is unavailable",
		fmt.Errorf("payer limit: %w", err),
	)
}
